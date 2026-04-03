package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// Sentinel errors for the project members repository.
var (
	// ErrProjectNotFound is returned when the requested project does not exist.
	ErrProjectNotFound = errors.New("project not found")

	// ErrMemberNotFound is returned when the requested project member does not exist.
	ErrMemberNotFound = errors.New("project member not found")

	// ErrUserNotFound is returned when the user lookup by email yields no result.
	ErrUserNotFound = errors.New("user not found")

	// ErrMemberAlreadyExists is returned when the user is already a member of the project.
	ErrMemberAlreadyExists = errors.New("user is already a member of this project")
)

var tracer = otel.Tracer("onboarding/repositories")

// PostgresProjectMembersRepository implements the persistence layer for projects and members.
type PostgresProjectMembersRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewProjectMembersRepository constructs a PostgresProjectMembersRepository.
func NewProjectMembersRepository(db *sql.DB, logger *zap.Logger) *PostgresProjectMembersRepository {
	return &PostgresProjectMembersRepository{db: db, logger: logger}
}

// CreateProject inserts a new project record and returns the created entity.
func (r *PostgresProjectMembersRepository) CreateProject(ctx context.Context, ownerID, name, projectType string) (*onboardingv1.Project, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.CreateProject")
	defer span.End()

	const q = `
		INSERT INTO projects (owner_id, name, type, created_at, updated_at)
		VALUES ($1, $2, $3::project_type, now(), now())
		RETURNING id, owner_id, name, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, q, ownerID, name, projectType)

	var (
		p         onboardingv1.Project
		createdAt string
		updatedAt string
	)
	if err := row.Scan(&p.Id, &p.OwnerId, &p.Name, &createdAt, &updatedAt); err != nil {
		r.logger.Error("project_members_repository.create_project: scan failed",
			zap.String("owner_id", ownerID),
			zap.Error(err))
		return nil, translateProjectMembersRepositoryError(err)
	}
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

// GetProject fetches a project by its ID.
func (r *PostgresProjectMembersRepository) GetProject(ctx context.Context, projectID string) (*onboardingv1.Project, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.GetProject")
	defer span.End()

	const q = `
		SELECT id, owner_id, name, created_at, updated_at
		FROM projects
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, q, projectID)
	var (
		p         onboardingv1.Project
		createdAt string
		updatedAt string
	)
	if err := row.Scan(&p.Id, &p.OwnerId, &p.Name, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NewCatalogError(apperrors.ErrProjectNotFound).WithError(ErrProjectNotFound)
		}
		r.logger.Error("project_members_repository.get_project: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, translateProjectMembersRepositoryError(err)
	}
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

// FindUserByEmail looks up a user ID by their email address.
func (r *PostgresProjectMembersRepository) FindUserByEmail(ctx context.Context, email string) (string, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.FindUserByEmail")
	defer span.End()

	const q = `SELECT id FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, q, email)

	var userID string
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperrors.NewCatalogError(apperrors.ErrResourceNotFound).WithError(ErrUserNotFound)
		}
		r.logger.Error("project_members_repository.find_user_by_email: query failed",
			zap.String("email", email),
			zap.Error(err))
		return "", translateProjectMembersRepositoryError(err)
	}
	return userID, nil
}

// CreateMember inserts a project membership record.
func (r *PostgresProjectMembersRepository) CreateMember(ctx context.Context, projectID, userID, invitedBy, role string) (*onboardingv1.ProjectMember, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.CreateMember")
	defer span.End()

	const q = `
		INSERT INTO project_members (project_id, user_id, role, invited_by, created_at, updated_at)
		VALUES ($1, $2, $3::project_member_role, $4, now(), now())
		RETURNING id, project_id, user_id, role::text, COALESCE(invited_by::text, ''), created_at, updated_at`

	var invitedByArg interface{} = invitedBy
	if invitedBy == "" {
		invitedByArg = nil
	}

	row := r.db.QueryRowContext(ctx, q, projectID, userID, role, invitedByArg)
	m, err := scanMember(row)
	if err != nil {
		if isDuplicateMemberError(err) {
			return nil, apperrors.NewCatalogError(apperrors.ErrResourceAlreadyExists).WithError(ErrMemberAlreadyExists)
		}
		r.logger.Error("project_members_repository.create_member: insert failed",
			zap.String("project_id", projectID),
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, translateProjectMembersRepositoryError(err)
	}
	return m, nil
}

// FindMemberByID fetches a single project member by memberID within a project.
func (r *PostgresProjectMembersRepository) FindMemberByID(ctx context.Context, projectID, memberID string) (*onboardingv1.ProjectMember, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.FindMemberByID")
	defer span.End()

	const q = `
		SELECT id, project_id, user_id, role::text, COALESCE(invited_by::text, ''), created_at, updated_at
		FROM project_members
		WHERE project_id = $1 AND id = $2`

	row := r.db.QueryRowContext(ctx, q, projectID, memberID)
	m, err := scanMember(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NewCatalogError(apperrors.ErrResourceNotFound).WithError(ErrMemberNotFound)
		}
		r.logger.Error("project_members_repository.find_member_by_id: query failed",
			zap.String("project_id", projectID),
			zap.String("member_id", memberID),
			zap.Error(err))
		return nil, translateProjectMembersRepositoryError(err)
	}
	return m, nil
}

// UpdateMemberRole updates the role of an existing member and returns the updated record.
func (r *PostgresProjectMembersRepository) UpdateMemberRole(ctx context.Context, projectID, memberID, newRole string) (*onboardingv1.ProjectMember, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.UpdateMemberRole")
	defer span.End()

	const q = `
		UPDATE project_members
		SET role = $3::project_member_role, updated_at = now()
		WHERE project_id = $1 AND id = $2
		RETURNING id, project_id, user_id, role::text, COALESCE(invited_by::text, ''), created_at, updated_at`

	row := r.db.QueryRowContext(ctx, q, projectID, memberID, newRole)
	m, err := scanMember(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NewCatalogError(apperrors.ErrResourceNotFound).WithError(ErrMemberNotFound)
		}
		r.logger.Error("project_members_repository.update_member_role: update failed",
			zap.String("project_id", projectID),
			zap.String("member_id", memberID),
			zap.Error(err))
		return nil, translateProjectMembersRepositoryError(err)
	}
	return m, nil
}

// ListMembers returns all members for the project with cursor-based pagination.
func (r *PostgresProjectMembersRepository) ListMembers(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*onboardingv1.ProjectMember, string, error) {
	ctx, span := tracer.Start(ctx, "ProjectMembersRepository.ListMembers")
	defer span.End()

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	const q = `
		SELECT id, project_id, user_id, role::text, COALESCE(invited_by::text, ''), created_at, updated_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY created_at ASC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, q, projectID, pageSize)
	if err != nil {
		r.logger.Error("project_members_repository.list_members: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, "", translateProjectMembersRepositoryError(err)
	}
	defer func() { _ = rows.Close() }()

	var members []*onboardingv1.ProjectMember
	for rows.Next() {
		var m onboardingv1.ProjectMember
		var (
			roleStr   string
			invitedBy string
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(&m.Id, &m.ProjectId, &m.UserId, &roleStr, &invitedBy, &createdAt, &updatedAt); err != nil {
			r.logger.Error("project_members_repository.list_members: scan failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, "", translateProjectMembersRepositoryError(err)
		}
		m.InvitedBy = invitedBy
		m.CreatedAt = createdAt
		m.UpdatedAt = updatedAt
		members = append(members, &m)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("project_members_repository.list_members: rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, "", translateProjectMembersRepositoryError(err)
	}

	nextToken := ""
	if int32(len(members)) == pageSize && len(members) > 0 {
		nextToken = members[len(members)-1].Id
	}

	return members, nextToken, nil
}

// scanMember scans a single project_members row.
func scanMember(row *sql.Row) (*onboardingv1.ProjectMember, error) {
	var (
		m         onboardingv1.ProjectMember
		roleStr   string
		invitedBy string
		createdAt string
		updatedAt string
	)
	if err := row.Scan(&m.Id, &m.ProjectId, &m.UserId, &roleStr, &invitedBy, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	m.InvitedBy = invitedBy
	m.CreatedAt = createdAt
	m.UpdatedAt = updatedAt
	return &m, nil
}

// isDuplicateMemberError detects PostgreSQL unique constraint violations for project_members.
func isDuplicateMemberError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "uq_project_members_project_user")
}

func translateProjectMembersRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	return apperrors.TranslateError(err, "repository")
}
