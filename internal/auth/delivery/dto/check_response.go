package dto

type IsExist struct {
	IsExist bool `json:"is_exist"`
}

type IsExistWithValidFlag struct {
	IsExist
	IsValidEmail bool `json:"is_valid"`
}
