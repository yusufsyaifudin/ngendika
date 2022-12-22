package msgsvc

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

type backendReport struct {
	BackendError  string // Errors some error occurred during job dispatching
	BackendReport *backend.Report
}

type senderWorkerJobReport struct {
	// Push notification provider id => report
	// One push notification provider may send and process multiple message,
	// so we use array as value to append the report under the same pnp id
	// Then we can map this into real value of PNP later
	BackendReports map[int64][]backendReport
}

type senderWorkerJob struct {
	Ctx  context.Context
	Lock sync.Locker
	Wg   *sync.WaitGroup

	SubmitTime      time.Time
	ServiceProvider backend.PushNotificationProvider
	Message         *backend.Message

	// Report must be pointer so we can append slice and read from the caller function.
	Report *senderWorkerJobReport
}

func senderWorker(workerID int, sender backend.SenderMux, jobs <-chan senderWorkerJob) {
	for job := range jobs {
		pnpID := job.ServiceProvider.ID

		job.Lock.Lock()
		if job.Report.BackendReports[pnpID] == nil {
			job.Report.BackendReports[pnpID] = make([]backendReport, 0)
		}

		job.Lock.Unlock()

		if job.Ctx == nil {
			job.Lock.Lock()

			job.Report.BackendReports[pnpID] = append(
				job.Report.BackendReports[pnpID],
				backendReport{
					BackendError: "no context passed in send fcm msg multicast job!",
				},
			)

			job.Lock.Unlock()
			job.Wg.Done()
			continue
		}

		if deadline, _ := job.Ctx.Deadline(); deadline.IsZero() {
			job.Lock.Lock()
			job.Report.BackendReports[pnpID] = append(
				job.Report.BackendReports[pnpID],
				backendReport{
					BackendError: "deadline context not defined in send fcm msg multicast job!",
				},
			)
			job.Lock.Unlock()
			job.Wg.Done()
			continue
		}

		var ctx = job.Ctx
		var span trace.Span
		ctx, span = tracer.StartSpan(job.Ctx, "msgsvc.senderWorker")
		span.SetAttributes(
			attribute.Int("worker_id", workerID),
			attribute.Int64("pnp_id", job.ServiceProvider.ID),
			attribute.String("pnp_label", job.ServiceProvider.Label),
		)

		report, err := sender.Send(ctx, workerID, job.ServiceProvider, job.Message)
		if err != nil {
			job.Lock.Lock()

			job.Report.BackendReports[pnpID] = append(
				job.Report.BackendReports[pnpID],
				backendReport{
					BackendError: fmt.Sprintf("error occured during send msg id '%s': %s", job.Message.ReferenceID, err),
				},
			)

			job.Lock.Unlock()

			span.End()
			job.Wg.Done()
			continue
		}

		job.Lock.Lock()
		if report != nil {
			job.Report.BackendReports[pnpID] = append(
				job.Report.BackendReports[pnpID],
				backendReport{
					BackendReport: report,
				},
			)
		}

		job.Lock.Unlock()

		span.End()
		job.Wg.Done()
		continue
	}

}
