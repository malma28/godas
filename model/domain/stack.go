package domain

type Stack struct {
	ID    string `json:"id" bson:"_id"`
	Items []Item `json:"items" bson:"items"`
	Owner string `json:"owner" bson:"owner"`
}
