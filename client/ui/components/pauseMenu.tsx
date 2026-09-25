import { useEffect, useRef } from "react";
import type Game from "../../game/game";
import type { RegistrationViewState } from "./onboardingOverlay";
import styles from "./pauseMenu.module.css";

type Props = {
  game: Game;
  registration: RegistrationViewState;
  authenticated: boolean;
  guest: boolean;
  onLogout: () => void;
  debugMode: boolean;
  onToggleDebugMode: () => void;
};

const buildSuffix = __BUILD_DIRTY__ ? "-dirty" : "";

export default function PauseMenu({ game, registration, authenticated, guest, onLogout, debugMode, onToggleDebugMode }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  const open = () => {
    dialogRef.current?.showModal();
    game.setMenuBlocked(true);
  };

  useEffect(() => {
    const dialog = dialogRef.current!;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.isComposing) {
        // Let the bank handle Escape for its item menu or immediate dismissal.
        if (!dialog.open && game.getBankPanelTargetId()) return;
        event.preventDefault();
        event.stopImmediatePropagation();
        if (event.repeat) return;
        if (dialog.open) {
          dialog.close();
        } else {
          dialog.showModal();
          game.setMenuBlocked(true);
        }
      }
    };
    const handleClose = () => game.setMenuBlocked(false);
    window.addEventListener("keydown", handleKeyDown, true);
    dialog.addEventListener("close", handleClose);
    return () => {
      window.removeEventListener("keydown", handleKeyDown, true);
      dialog.removeEventListener("close", handleClose);
      dialog.close();
      game.setMenuBlocked(false);
    };
  }, [game]);

  useEffect(() => {
    dialogRef.current?.close();
  }, [registration.phase, registration.idleWarningUntil]);

  return (
    <>
      <button
        className={styles.trigger}
        type="button"
        aria-label="Open menu"
        aria-haspopup="dialog"
        aria-keyshortcuts="Escape"
        onPointerDown={(event) => event.stopPropagation()}
        onPointerUp={(event) => event.stopPropagation()}
        onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); }}
        onClick={open}
        title="Open menu (Esc)"
      >
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
          <path d="M3 5h14M3 10h14M3 15h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </button>
      <dialog
        ref={dialogRef}
        className={styles.dialog}
        aria-labelledby="pause-menu-title"
        onKeyDown={(event) => event.stopPropagation()}
        onPointerDown={(event) => event.stopPropagation()}
        onPointerUp={(event) => event.stopPropagation()}
        onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); }}
        onClick={(event) => {
          event.stopPropagation();
          if (event.target === event.currentTarget) event.currentTarget.close();
        }}
      >
        <div className={styles.content}>
          <h1 id="pause-menu-title">Menu</h1>
          <button className={styles.resume} type="button" autoFocus onClick={() => dialogRef.current?.close()}>
            Return to game
          </button>
          <button className={styles.action} type="button" aria-pressed={debugMode} onClick={onToggleDebugMode}>
            {debugMode ? "Disable debug mode" : "Enable debug mode"}
          </button>
          <a
            className={styles.action}
            href="https://github.com/dazkins/webscape"
            target="_blank"
            rel="noreferrer"
            title="Webscape source code and AGPL-3.0-only license (provided without warranty)"
          >
            Source &amp; license (AGPLv3)
          </a>
          {authenticated && (
            <button className={styles.action} type="button" onClick={onLogout} disabled={registration.phase === "signingOut"}>
              {guest ? "End guest session" : "Sign out"}
            </button>
          )}
          <span className={styles.build} title={`Client build ${__BUILD_REVISION__}${buildSuffix}`}>
            Build {__BUILD_REVISION__.slice(0, 8)}{buildSuffix}
          </span>
        </div>
      </dialog>
    </>
  );
}
