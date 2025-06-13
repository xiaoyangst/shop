package response

import "time"

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	t := time.Time(j)
	return []byte(`"` + t.Format("2006-01-02") + `"`), nil
}

type UserResp struct {
	Id       int64    `json:"id"`
	NickName string   `json:"nickname"`
	Birthday JsonTime `json:"birthday"`
	Gender   string   `json:"gender"`
	Mobile   string   `json:"mobile"`
}
