import { type FormEvent, useEffect, useRef, useState } from "react";
import styles from "./onboardingOverlay.module.css";

export type RegistrationPhase =
  | "inactive"
  | "replaced"
  | "activeElsewhere"
  | "signedOut"
  | "signingOut"
  | "connectionError"
  | "connecting"
  | "nameEntry"
  | "registering"
  | "reconnecting"
  | "registered";

export type RegistrationViewState = {
  idleWarningUntil?: number;
  phase: RegistrationPhase;
  name: string;
  error: string;
};

type Props = {
  allowGuests: boolean;
  signInEnabled: boolean;
  onGuest: () => void;
  state: RegistrationViewState;
  onRegister: (name: string) => void;
  onLogin: () => void;
  onRetry: () => void;
  onStayConnected: () => void;
};

function validateName(name: string): string {
  const normalized = name.trim();
  if (normalized.length === 0) {
    return "Please enter a name.";
  }
  if (Array.from(normalized).length > 24) {
    return "Your name must be 24 characters or fewer.";
  }
  return "";
}

export default function OnboardingOverlay({ state, allowGuests, signInEnabled, onGuest, onRegister, onLogin, onRetry, onStayConnected }: Props) {
  const [name, setName] = useState(state.name);
  const [validationError, setValidationError] = useState("");
  const [serverError, setServerError] = useState(state.error);
  const [now, setNow] = useState(Date.now());
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setName(state.name);
  }, [state.name]);

  useEffect(() => {
    setServerError(state.error);
  }, [state.error]);

  useEffect(() => {
    if (state.phase === "nameEntry") {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [state.phase]);

  useEffect(() => {
    if (!state.idleWarningUntil) return;
    setNow(Date.now());
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [state.idleWarningUntil]);

  if (state.phase === "registered") {
    if (!state.idleWarningUntil) return null;
    const seconds = Math.max(0, Math.ceil((state.idleWarningUntil - now) / 1000));
    return (
      <div className={styles.warning} data-inactivity-warning role="region" aria-label="Inactivity warning"
        onPointerDown={(event) => event.stopPropagation()} onClick={(event) => event.stopPropagation()}>
        <p role="status">You’ll be disconnected for inactivity in {seconds} seconds.</p>
        <button type="button" onClick={onStayConnected}>Stay connected</button>
      </div>
    );
  }

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    const error = validateName(name);
    setValidationError(error);
    if (!error) {
      onRegister(name);
    }
  };

  const error = validationError || serverError;
  const isNameEntry = state.phase === "nameEntry";

  return (
    <div className={styles.backdrop} role="presentation">
      <section
        className={styles.card}
        role="dialog"
        aria-modal="true"
        aria-labelledby="onboarding-title"
        aria-describedby="onboarding-description"
      >
        {state.phase === "inactive" || state.phase === "replaced" || state.phase === "activeElsewhere" ? (
          <div>
            <h1 id="onboarding-title">{state.phase === "inactive" ? "Disconnected for inactivity" : state.phase === "replaced" ? "Your game was opened elsewhere" : "Already playing elsewhere"}</h1>
            <p id="onboarding-description" className={styles.description}>
              {state.phase === "inactive" ? "You’re still signed in. Rejoin whenever you’re ready." : state.phase === "replaced" ? "This game connection has ended. You’re still signed in." : "Continue here to end your other game connection and bring your character to this device."}
            </p>
            <button type="button" onClick={onRetry}>{state.phase === "activeElsewhere" ? "Continue here" : "Rejoin game"}</button>
          </div>
        ) : state.phase === "signedOut" || state.phase === "connectionError" ? (
          <div>
            <h1 id="onboarding-title">{state.phase === "signedOut" ? "Your adventure awaits" : "The path is interrupted"}</h1>
            <p id="onboarding-description" className={styles.description}>
              {state.phase === "signedOut" ? (allowGuests ? (signInEnabled ? "Sign in to return to your character, or play as a guest. Guest progress is not saved across sessions." : "Play without an account. Guest progress is not saved across sessions.") : "Sign in to enter the world and return to your character.") : "Check your connection, then try again."}
            </p>
            {state.error && <p role="alert" className={styles.error}>{state.error}</p>}
            {state.phase === "signedOut" ? (
              <div className={styles.loginActions}>
                {signInEnabled && <button type="button" onClick={onLogin}>Sign in</button>}
                {allowGuests && <button type="button" onClick={onGuest}>Play as guest</button>}
              </div>
            ) : <button type="button" onClick={onRetry}>Reconnect</button>}
          </div>
        ) : isNameEntry ? (
          <form onSubmit={handleSubmit} noValidate>
            <h1 id="onboarding-title">What’s your name, adventurer?</h1>
            <p id="onboarding-description" className={styles.description}>
              Choose the name other travelers will see in the world.
            </p>
            <label className={styles.label} htmlFor="adventurer-name">
              Display name
            </label>
            <input
              ref={inputRef}
              id="adventurer-name"
              className={error ? styles.inputError : ""}
              type="text"
              value={name}
              maxLength={48}
              autoComplete="nickname"
              enterKeyHint="go"
              aria-invalid={Boolean(error)}
              aria-describedby={error ? "name-error" : "name-hint"}
              onChange={(event) => {
                setName(event.target.value);
                setValidationError("");
                setServerError("");
              }}
            />
            <div className={styles.feedback} aria-live="polite">
              {error ? (
                <span id="name-error" className={styles.error}>{error}</span>
              ) : (
                <span id="name-hint">1–24 characters</span>
              )}
            </div>
            <button type="submit">Start my adventure</button>
          </form>
        ) : (
          <div className={styles.status} aria-live="polite">
            <div className={styles.spinner} aria-hidden="true" />
            <h1 id="onboarding-title">
              {state.phase === "signingOut" ? "Signing out…" : state.phase === "reconnecting"
                ? "Finding the path back…"
                : state.phase === "connecting"
                  ? "Opening the way…"
                  : "Entering the world…"}
            </h1>
            <p id="onboarding-description" className={styles.description}>
              {state.phase === "reconnecting"
                ? "Your connection was interrupted. Rejoining automatically."
                : state.name
                  ? `Preparing your adventure, ${state.name}.`
                  : "Connecting to the realm."}
            </p>
          </div>
        )}
      </section>
    </div>
  );
}
