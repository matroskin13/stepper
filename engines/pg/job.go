package pg

type Job struct {
	Status       string `json:"status"`
	Name         string `json:"name"`
	Pattern      string `json:"pattern"`
	NextLaunchAt int64  `json:"naxtLaunchAt"`
}
