package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/api"
	"github.com/sable-inc/anvil/internal/output"
)

// VideoJob represents a video processing job status response.
type VideoJob struct {
	JobID    string          `json:"jobId" yaml:"jobId"`
	Status   string          `json:"status" yaml:"status"`
	Progress *VideoProgress  `json:"progress,omitempty" yaml:"progress,omitempty"`
	Result   json.RawMessage `json:"result,omitempty" yaml:"result,omitempty"`
	Error    *VideoError     `json:"error,omitempty" yaml:"error,omitempty"`
}

// VideoProgress describes the current stage of a video processing job.
type VideoProgress struct {
	Stage        string `json:"stage" yaml:"stage"`
	Progress     int    `json:"progress" yaml:"progress"`
	Message      string `json:"message" yaml:"message"`
	CurrentChunk *int   `json:"currentChunk,omitempty" yaml:"currentChunk,omitempty"`
	TotalChunks  *int   `json:"totalChunks,omitempty" yaml:"totalChunks,omitempty"`
	CurrentStage *int   `json:"currentStage,omitempty" yaml:"currentStage,omitempty"`
	TotalStages  *int   `json:"totalStages,omitempty" yaml:"totalStages,omitempty"`
}

// VideoError describes an error from video processing.
type VideoError struct {
	Code      string `json:"code" yaml:"code"`
	Message   string `json:"message" yaml:"message"`
	Retryable bool   `json:"retryable" yaml:"retryable"`
}

func newVideoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Video processing",
		Long:  "Generate moments and journeys from video, and track job status.",
	}

	cmd.AddCommand(newVideoGenerateMomentCmd())
	cmd.AddCommand(newVideoGenerateJourneyCmd())
	cmd.AddCommand(newVideoJobStatusCmd())
	return cmd
}

func newVideoGenerateMomentCmd() *cobra.Command {
	var (
		videoURL          string
		filename          string
		workflowDesc      string
		productName       string
		targetAudience    string
		notes             string
		watch             bool
	)

	cmd := &cobra.Command{
		Use:   "generate-moment",
		Short: "Start a moment generation job from video",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"videoUrl": videoURL}
			if filename != "" {
				body["filename"] = filename
			}

			ctx := buildAdditionalContext(workflowDesc, productName, targetAudience, notes)
			if len(ctx) > 0 {
				body["additionalContext"] = ctx
			}

			var resp struct {
				JobID  string `json:"jobId"`
				Status string `json:"status"`
			}
			if err := client.Post(cmd.Context(), "/video-processing/moment/start", body, &resp); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(a.Out, "Job started: %s (status: %s)\n", resp.JobID, resp.Status)

			if !watch {
				return nil
			}

			return pollVideoJob(cmd.Context(), a.Out, client, resp.JobID)
		},
	}

	cmd.Flags().StringVar(&videoURL, "video-url", "", "URL of the video to process (required)")
	cmd.Flags().StringVar(&filename, "filename", "", "Filename hint")
	cmd.Flags().StringVar(&workflowDesc, "workflow-description", "", "Description of the workflow shown")
	cmd.Flags().StringVar(&productName, "product-name", "", "Product name for context")
	cmd.Flags().StringVar(&targetAudience, "target-audience", "", "Target audience")
	cmd.Flags().StringVar(&notes, "notes", "", "Additional notes")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until job completes")
	_ = cmd.MarkFlagRequired("video-url")
	return cmd
}

func newVideoGenerateJourneyCmd() *cobra.Command {
	var (
		videoURL   string
		filename   string
		stageNames string
		genContext string
		mode       string
		watch      bool
	)

	cmd := &cobra.Command{
		Use:   "generate-journey",
		Short: "Start a journey generation job from video",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"videoUrl": videoURL}
			if filename != "" {
				body["filename"] = filename
			}
			if mode != "" {
				body["mode"] = mode
			}

			hints := map[string]any{}
			if stageNames != "" {
				hints["generalContext"] = stageNames
			}
			if genContext != "" {
				hints["generalContext"] = genContext
			}
			if len(hints) > 0 {
				body["stageHints"] = hints
			}

			var resp struct {
				JobID  string `json:"jobId"`
				Status string `json:"status"`
			}
			if err := client.Post(cmd.Context(), "/video-processing/journey/start", body, &resp); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(a.Out, "Job started: %s (status: %s)\n", resp.JobID, resp.Status)

			if !watch {
				return nil
			}

			return pollVideoJob(cmd.Context(), a.Out, client, resp.JobID)
		},
	}

	cmd.Flags().StringVar(&videoURL, "video-url", "", "URL of the video to process (required)")
	cmd.Flags().StringVar(&filename, "filename", "", "Filename hint")
	cmd.Flags().StringVar(&stageNames, "stage-names", "", "Comma-separated stage name hints")
	cmd.Flags().StringVar(&genContext, "general-context", "", "General context about the video")
	cmd.Flags().StringVar(&mode, "mode", "", "Mode: post-sales or pre-sales")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until job completes")
	_ = cmd.MarkFlagRequired("video-url")
	return cmd
}

func newVideoJobStatusCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "job-status <jobId>",
		Short: "Get video processing job status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var job VideoJob
			if err := client.Get(cmd.Context(), "/video-processing/jobs/"+args[0], &job); err != nil {
				return err
			}

			if !watch {
				return output.Write(a.Out, a.Format, job, videoJobTable(job))
			}

			// Already terminal?
			if job.Status == "completed" || job.Status == "error" {
				return output.Write(a.Out, a.Format, job, videoJobTable(job))
			}

			return pollVideoJob(cmd.Context(), a.Out, client, args[0])
		},
	}

	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until job completes")
	return cmd
}

func videoJobTable(job VideoJob) *output.Table {
	t := output.NewTable("Field", "Value")
	t.AddRow("Job ID", job.JobID)
	t.AddRow("Status", job.Status)
	if job.Progress != nil {
		t.AddRow("Stage", job.Progress.Stage)
		t.AddRow("Progress", fmt.Sprintf("%d%%", job.Progress.Progress))
		if job.Progress.Message != "" {
			t.AddRow("Message", job.Progress.Message)
		}
	}
	if job.Error != nil {
		t.AddRow("Error Code", job.Error.Code)
		t.AddRow("Error", job.Error.Message)
	}
	return t
}

func pollVideoJob(ctx context.Context, w io.Writer, client *api.Client, jobID string) error {
	var lastStatus string
	return output.Poll(ctx, w, output.PollConfig{
		Interval: 2 * time.Second,
		Timeout:  30 * time.Minute,
		StatusFunc: func(ctx context.Context) (string, bool, error) {
			var job VideoJob
			if err := client.Get(ctx, "/video-processing/jobs/"+jobID, &job); err != nil {
				return "", false, err
			}
			status := job.Status
			if job.Progress != nil {
				status = fmt.Sprintf("%s (%s %d%%)", job.Status, job.Progress.Stage, job.Progress.Progress)
			}
			done := job.Status == "completed" || job.Status == "error"
			return status, done, nil
		},
		OnStatus: func(status string) {
			if status != lastStatus {
				_, _ = fmt.Fprintf(w, "  %s\n", status)
				lastStatus = status
			}
		},
	})
}

func buildAdditionalContext(workflowDesc, productName, targetAudience, notes string) map[string]any {
	ctx := map[string]any{}
	if workflowDesc != "" {
		ctx["workflowDescription"] = workflowDesc
	}
	if productName != "" {
		ctx["productName"] = productName
	}
	if targetAudience != "" {
		ctx["targetAudience"] = targetAudience
	}
	if notes != "" {
		ctx["notes"] = notes
	}
	return ctx
}
