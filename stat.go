package stat

type Stat struct {
	SiteId    int64
	Rank      int64
	TodayHit  int64
	TodayUniq int64
}

func NewStat() *Stat {
	return &Stat{}
}
