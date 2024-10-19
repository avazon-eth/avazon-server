package dto

type AvatarCreationRequest struct {
	Name        string `json:"name" binding:"required,notempty"`        // ex) John Doe
	Species     string `json:"species" binding:"required,notempty"`     // ex) human, alien, robot, etc.
	Gender      string `json:"gender" binding:"required,notempty"`      // ex) male, female, mixed, etc.
	Age         int    `json:"age" binding:"required,gte=1"`            // Age must be greater than 0
	Language    string `json:"language" binding:"required,notempty"`    // ex) English
	Country     string `json:"country" binding:"required,notempty"`     // ex) United States
	ImageStyle  string `json:"image_style" binding:"required,notempty"` // cartoon, realistic
	Description string `json:"description"`                             // optional
}
