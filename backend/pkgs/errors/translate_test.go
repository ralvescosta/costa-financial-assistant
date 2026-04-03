package errors

import (
	nativeerrors "errors"
	"testing"
)

func TestTranslateErrorUnknownFallback_T063(t *testing.T) {
	nativeErr := nativeerrors.New("custom-native-error")
	appErr := TranslateError(nativeErr, "unmapped_layer")
	if appErr == nil {
		t.Fatalf("expected non-nil AppError")
	}
	if appErr.Code != ErrUnknown.Code {
		t.Fatalf("expected unknown fallback code %s, got %s", ErrUnknown.Code, appErr.Code)
	}
}

func TestTranslateErrorNilNative_T063(t *testing.T) {
	if TranslateError(nil, "service") != nil {
		t.Fatalf("expected nil translation for nil native error")
	}
}

func TestAsAppErrorCompatibility_T063(t *testing.T) {
	nativeErr := nativeerrors.New("native")
	appErr := NewCatalogError(ErrDatabaseError).WithError(nativeErr)

	if !nativeerrors.Is(appErr, nativeErr) {
		t.Fatalf("errors.Is compatibility expected")
	}
	if AsAppError(appErr) == nil {
		t.Fatalf("AsAppError should unwrap AppError")
	}
}
