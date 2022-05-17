package utils

import (
	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	//st.MachineID = awsutil.AmazonEC2MachineID
	sf = sonyflake.NewSonyflake(sonyflake.Settings{})
	if sf == nil {
		panic("sonyflake not created")
	}
}

func GetUUID() (uint64, error) {
	return sf.NextID()
}
