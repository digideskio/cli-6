package worker

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/daticahealth/cli/commands/services"
	"github.com/daticahealth/cli/lib/jobs"
	"github.com/daticahealth/cli/models"
	"github.com/olekukonko/tablewriter"
)

func CmdList(svcName string, iw IWorker, is services.IServices, ij jobs.IJobs) error {
	service, err := is.RetrieveByLabel(svcName)
	if err != nil {
		return err
	}
	if service == nil {
		return fmt.Errorf("Could not find a service with the label \"%s\". You can list services with the \"datica services list\" command.", svcName)
	}
	workers, err := iw.Retrieve(service.ID)
	if err != nil {
		return err
	}

	jobs, err := ij.RetrieveByType(service.ID, "worker", 1, 1000)
	if err != nil {
		return err
	}
	type workerJob struct {
		scale   int
		running int
	}
	var workerJobs = map[string]*workerJob{}
	for target, scale := range workers.Workers {
		workerJobs[target] = &workerJob{scale, 0}
	}
	if len(workerJobs) == 0 {
		logrus.Printf("No running workers found for %s", svcName)
		logrus.Printf("\nYou are using 0 out of your available %d workers for %s", service.WorkerScale, svcName)
		return nil
	}
	for _, j := range *jobs {
		if _, ok := workerJobs[j.Target]; !ok {
			workerJobs[j.Target] = &workerJob{0, 0}
		}
		if j.Status == "running" {
			workerJobs[j.Target].running += 1
		}
	}

	data := [][]string{{"TARGET", "SCALE", "RUNNING JOBS"}}
	total := 0
	for target, wj := range workerJobs {
		total += wj.scale
		data = append(data, []string{target, fmt.Sprintf("%d", wj.scale), fmt.Sprintf("%d", wj.running)})
	}

	table := tablewriter.NewWriter(logrus.StandardLogger().Out)
	table.SetBorder(false)
	table.SetRowLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.AppendBulk(data)
	table.Render()
	logrus.Printf("\nYou are using %d out of your available %d workers for %s", total, service.WorkerScale, svcName)
	return nil
}

func (w *SWorker) Retrieve(svcID string) (*models.Workers, error) {
	headers := w.Settings.HTTPManager.GetHeaders(w.Settings.SessionToken, w.Settings.Version, w.Settings.Pod, w.Settings.UsersID)
	resp, statusCode, err := w.Settings.HTTPManager.Get(nil, fmt.Sprintf("%s%s/environments/%s/services/%s/workers", w.Settings.PaasHost, w.Settings.PaasHostVersion, w.Settings.EnvironmentID, svcID), headers)
	if err != nil {
		return nil, err
	}
	var workers models.Workers
	err = w.Settings.HTTPManager.ConvertResp(resp, statusCode, &workers)
	if err != nil {
		return nil, err
	}
	return &workers, nil
}
