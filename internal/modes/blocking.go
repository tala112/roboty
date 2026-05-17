package modes

import (
	"log"
	"time"
)

const killTimeout = 10 * time.Second

var globalBlockingVerifier = GetGlobalSafetyVerifier()

func CloseApps(apps []string) {
	killer := NewRealProcessKiller()
	for _, app := range apps {
		safeExec := safeExecName(app)
		if safeExec == "" {
			continue
		}
		if IsDevMode() {
			log.Printf("[dev] WOULD kill %s", safeExec)
			continue
		}
		if err := killer.Kill(safeExec, killTimeout); err != nil {
			log.Printf("[blocking] close %s: %v", safeExec, err)
		}
	}
}

func CloseApp(appExec string) {
	CloseApps([]string{appExec})
}

func safeExecName(app string) string {
	safeExec := NormalizeKillExec(app)
	if safeExec == "" {
		log.Printf("[blocking] SKIP kill of %q: rejected by NormalizeKillExec", app)
		return ""
	}
	safe, reason := globalBlockingVerifier.IsSafeToKill(safeExec)
	if !safe {
		log.Printf("[blocking] SAFETY BLOCKED kill of %q: %s", app, reason)
		return ""
	}
	return safeExec
}

func isAppRunning(execName string) bool {
	killer := NewRealProcessKiller()
	ok, err := killer.IsRunning(execName)
	return err == nil && ok
}
