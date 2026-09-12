import { type FormEvent, useEffect, useRef, useState } from "react";
import styles from "./onboardingOverlay.module.css";

export type RegistrationPhase =
  | "signedOut"
  | "signingOut"
  | "connectionError"
  | "connecting"
  | "nameEntry"
  | "registering"
  | "reconnecting"
  | "registered";

export type RegistrationViewState = {
  phase: RegistrationPhase;
  name: string;
  error: string;
};

type Props = {
  guest: boolean;
  state: RegistrationViewState;
  onRegister: (name: string) => void;
  onLogin: () => void;
  onRetry: () => void;
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

export default function OnboardingOverlay({ state, guest, onRegister, onLogin, onRetry }: Props) {
  const [name, setName] = useState(state.name);
  const [validationError, setValidationError] = useState("");
  const [serverError, setServerError] = useState(state.error);
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

  if (state.phase === "registered") {
    return null;
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
        {state.phase === "signedOut" || state.phase === "connectionError" ? (
          <div>
            <h1 id="onboarding-title">{state.phase === "signedOut" ? "Your adventure awaits" : "The path is interrupted"}</h1>
            <p id="onboarding-description" className={styles.description}>
              {state.phase === "signedOut" ? (guest ? "Play without an account. Your guest character lasts for this session." : "Sign in to enter the world and return to your character.") : "Check your connection, then try again."}
            </p>
            {state.error && <p role="alert" className={styles.error}>{state.error}</p>}
            <button type="button" onClick={state.phase === "signedOut" ? onLogin : onRetry}>
              {state.phase === "signedOut" ? (guest ? "Play as guest" : "Sign in") : "Try again"}
            </button>
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
