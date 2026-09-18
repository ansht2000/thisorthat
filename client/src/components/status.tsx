import type { ReactNode } from "react";
import "./status.css";

export function LoadingState({ label = "Loading" }: { label?: string }) {
    return (
        <div className="state" role="status">
            <span className="stateSpinner" aria-hidden="true" />
            <span className="stateMessage">{label}…</span>
        </div>
    );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
    return (
        <div className="state" role="alert">
            <p className="stateTitle">Something went wrong</p>
            <p className="stateMessage">{message}</p>
            {onRetry && (
                <button type="button" className="stateButton" onClick={onRetry}>
                    Try again
                </button>
            )}
        </div>
    );
}

export function EmptyState({ title, message, children }: { title: string; message?: string; children?: ReactNode }) {
    return (
        <div className="state">
            <p className="stateTitle">{title}</p>
            {message && <p className="stateMessage">{message}</p>}
            {children}
        </div>
    );
}
