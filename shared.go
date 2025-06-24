package temporaltalk

// a queue of tasks that workers poll constantly
// 3 types: workflow, activity and nexus
var TaskQueueName = "temporaltalk"

type WeatherOutput struct {
	Name            string  `json:"name"`
	IpAddress       string  `json:"ip_address"`
	City            string  `json:"city"`
	CurrentForecast float32 `json:"current_forecast"`
}

type Geopoint struct {
	City      string  `json:"city"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
}

type WorkflowInput struct {
	Name   string `json:"name"`
	Moving bool   `json:"moving"`
}
