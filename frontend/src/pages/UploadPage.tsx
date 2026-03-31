import { useRef, useState } from 'react'
import { useUploadDocument } from '@/hooks/useUploadDocument'
import { useDocuments } from '@/hooks/useDocuments'
import type { DocumentKind } from '@/types/documents'

const KIND_OPTIONS: { label: string; value: DocumentKind }[] = [
  { label: 'Auto-detect', value: 'unspecified' },
  { label: 'Bill', value: 'bill' },
  { label: 'Statement', value: 'statement' },
]

/** UploadPage — allows uploading a PDF document and lists project-scoped documents. */
export default function UploadPage() {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [selectedKind, setSelectedKind] = useState<DocumentKind>('unspecified')
  const [feedbackMsg, setFeedbackMsg] = useState<string | null>(null)
  const [isErrorFeedback, setIsErrorFeedback] = useState(false)

  const { mutate: upload, isPending } = useUploadDocument({
    kind: selectedKind,
    onSuccess: (doc) => {
      setFeedbackMsg(`Uploaded "${doc.fileName}" successfully.`)
      setIsErrorFeedback(false)
    },
    onError: (err) => {
      setFeedbackMsg(err.message)
      setIsErrorFeedback(true)
    },
  })

  const { documents, isLoading, isError } = useDocuments()

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setFeedbackMsg(null)
    upload(file)
    // Reset so the same file can be re-selected after an error.
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-semibold text-text-primary">
        Upload Document
      </h1>

      {/* ── Upload zone ──────────────────────────────────────────────── */}
      <section
        aria-label="Upload area"
        className="mb-6 rounded-lg border-2 border-dashed border-border bg-surface-raised p-8 text-center"
      >
        <p className="mb-4 text-sm text-text-secondary">
          Select a PDF file to upload
        </p>

        <div className="mb-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <select
            value={selectedKind}
            onChange={(e) => setSelectedKind(e.target.value as DocumentKind)}
            className="rounded border border-border bg-surface px-3 py-1.5 text-sm text-text-primary"
            aria-label="Document kind"
          >
            {KIND_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>

          <label
            htmlFor="file-input"
            className={`cursor-pointer rounded-md px-4 py-2 text-sm font-medium text-white transition-colors ${isPending
                ? 'cursor-not-allowed bg-primary opacity-60'
                : 'bg-primary hover:bg-primary-hover'
              }`}
          >
            {isPending ? 'Uploading…' : 'Choose File'}
          </label>
          <input
            id="file-input"
            ref={fileInputRef}
            type="file"
            accept="application/pdf"
            className="sr-only"
            disabled={isPending}
            onChange={handleFileChange}
          />
        </div>

        {feedbackMsg && (
          <p
            role="status"
            className={`text-sm ${isErrorFeedback ? 'text-danger' : 'text-success'}`}
          >
            {feedbackMsg}
          </p>
        )}
      </section>

      {/* ── Documents list ───────────────────────────────────────────── */}
      <section aria-label="Uploaded documents">
        <h2 className="mb-3 text-lg font-medium text-text-primary">
          Your Documents
        </h2>

        {isLoading && (
          <p className="text-sm text-text-secondary">Loading documents…</p>
        )}

        {isError && (
          <p className="text-sm text-danger">Failed to load documents.</p>
        )}

        {!isLoading && !isError && documents.length === 0 && (
          <p className="text-sm text-text-secondary">
            No documents uploaded yet.
          </p>
        )}

        {documents.length > 0 && (
          <ul className="divide-y divide-border rounded-lg border border-border bg-surface">
            {documents.map((doc) => (
              <li
                key={doc.id}
                className="flex items-center justify-between px-4 py-3"
              >
                <div>
                  <p className="text-sm font-medium text-text-primary">
                    {doc.fileName}
                  </p>
                  <p className="text-xs text-text-secondary">
                    {doc.kind !== 'unspecified' ? doc.kind : 'unclassified'} ·{' '}
                    {doc.analysisStatus}
                  </p>
                </div>
                <span className="text-xs text-text-secondary">
                  {new Date(doc.uploadedAt).toLocaleDateString()}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  )
}
