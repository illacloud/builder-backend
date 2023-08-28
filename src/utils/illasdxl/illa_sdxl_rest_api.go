package illasdxl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
)

const (
	// api route part
	CREATE_PREDICTION = "/predictions"

	NEGATIVE_PROMPT = "multiple objects"
	DEFAULT_WIDTH   = 512
	DEFAULT_HEIGHT  = 512

	PROMPT_PREFIX = "illustration, 200*200, made up of a single object and solid color background, the object must be associated with the sentences: "
)

type IllaSDXLRestAPI struct {
	Config *config.Config
}

type RawInput struct {
	Prompt            string  `json:"prompt,omitempty"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Width             int     `json:"width,omitempty"`
	Height            int     `json:"height,omitempty"`
	NumOutputs        int     `json:"num_outputs,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
	Scheduler         string  `json:"scheduler,omitempty"`
	Seed              int     `json:"seed,omitempty"`
}

type PredictionInput struct {
	Input RawInput `json:"input"`
}

type PredictionOutput struct {
	Input       RawInput  `json:"input"`
	Output      []string  `json:"output,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Logs        string    `json:"logs,omitempty"`
	Error       string    `json:"error,omitempty"`
	Status      string    `json:"status,omitempty"`
	Metrics     struct {
		PredictTime float64 `json:"predict_time,omitempty"`
	} `json:"metrics"`
}

func CreatePrediction(input string) (string, error) {
	prompt := PROMPT_PREFIX + input
	cfg := config.GetInstance()
	client := resty.New()
	clientTimeout, err := time.ParseDuration(cfg.SDXLAPITimeout)
	if err != nil {
		clientTimeout = 5 * time.Minute
	}
	client.SetTimeout(clientTimeout)
	resp, err := client.R().
		SetHeader("Authorization", cfg.SDXLAPIKey).
		SetBody(PredictionInput{Input: RawInput{Prompt: prompt, Width: DEFAULT_WIDTH, Height: DEFAULT_HEIGHT}}).
		Post(cfg.SDXLAPI + fmt.Sprintf(CREATE_PREDICTION))
	fmt.Printf("[IllaSDXLRestAPI.CreatePrediction()]  response status code: %d, err: %+v", resp.StatusCode(), err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", err
		}
		return "", errors.New("the queue has reached maximum capacity")
	}
	var output PredictionOutput
	errInUnmarshal := json.Unmarshal([]byte(resp.String()), &output)
	if errInUnmarshal != nil {
		return "", errInUnmarshal
	}
	if len(output.Output) != 1 {
		return "", errors.New("NSFW content detected. Try running it again, or try a different prompt")
	}
	imgURL := output.Output[0]
	return imgURL, nil
}
