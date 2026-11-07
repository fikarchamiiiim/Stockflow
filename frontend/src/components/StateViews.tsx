type ErrorStateProps = {
  message: string
  onRetry: () => void
}

export function LoadingState() {
  return <p className="state muted">loading...</p>
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="state state-error" role="alert">
      <p>Failed to load data: {message}</p>
      <button type="button" onClick={onRetry}>
        retry
      </button>
    </div>
  )
}

export function EmptyState({ message }: { message: string }) {
  return <p className="state muted">{message}</p>
}
