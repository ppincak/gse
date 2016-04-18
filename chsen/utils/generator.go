package utils

import (
	"github.com/nu7hatch/gouuid"
)

func GenerateUID() string {
	uid, err := uuid.NewV4()
	if err != nil{

	}
	return uid.String()
}