package data

import (
	"farentalapp/core/data/api"
)

type StatCode string

const (
	HPStat  StatCode = "hp"
	MPStat  StatCode = "mp"
	INIStat StatCode = "ini"
	STRStat StatCode = "str"
	INTStat StatCode = "int"
	LUKStat StatCode = "luk"
	PREStat StatCode = "pre"
	AGIStat StatCode = "agi"
	DEFStat StatCode = "def"
	MDEStat StatCode = "mde"
	ATKStat StatCode = "atk"
)

type Stat struct {
	Value    int
	MaxValue int
}

type StatMap map[StatCode]Stat

func NewStatMap(stats []api.CharacterStatResponse) StatMap {
	s := StatMap{}

	for _, stat := range stats {
		st := Stat{}

		st.Value = stat.Value
		st.MaxValue = stat.MaxValue

		s[StatCode(stat.Code)] = st
	}

	return s
}
