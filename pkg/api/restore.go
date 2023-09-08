package api

import (
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/config"
	"github.com/stefanprodan/mgob/pkg/notifier"
	"github.com/stefanprodan/mgob/pkg/restore"
)

func postRestore(w http.ResponseWriter, r *http.Request) {
	cfg := r.Context().Value("app.config").(config.AppConfig)
	modules := r.Context().Value("app.modules").(config.ModuleConfig)
	planID := chi.URLParam(r, "planID")
	backupPath := chi.URLParam(r, "backupPath")
	plan, err := config.LoadPlan(cfg.ConfigPath, planID)
	if err != nil {
		render.Status(r, 500)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	log.WithField("plan", planID).Infof("On demand restore started from %v", backupPath)

	res, err := restore.Run(plan, &cfg, &modules, backupPath)
	if err != nil {
		log.WithField("plan", planID).Errorf("On demand restore failed %v", err)
		if err := notifier.SendNotification(fmt.Sprintf("RESTORE FAILED: %v on demand restore failed", planID),
			err.Error(), true, plan); err != nil {
			log.WithField("plan", plan.Name).Errorf("Notifier failed for on demand restore %v", err)
		}
		render.Status(r, 500)
		render.JSON(w, r, map[string]string{"error": err.Error()})
	} else {
		log.WithField("plan", plan.Name).Infof("On demand restore finished in %v, restore from %v size %v",
			res.Duration, res.Name, humanize.Bytes(uint64(res.Size)))
		if err := notifier.SendNotification(fmt.Sprintf("%v on demand restore finished", plan.Name),
			fmt.Sprintf("%v restore finished in %v archive size %v",
				res.Name, res.Duration, humanize.Bytes(uint64(res.Size))),
			false, plan); err != nil {
			log.WithField("plan", plan.Name).Errorf("Notifier failed for on demand restore %v", err)
		}
		render.JSON(w, r, toBackupResult(res))
	}
}
