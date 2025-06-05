package api

type CraftStartBody struct {
	RecipeID uint `validate:"required"`
	Amount   int  `validate:"required,number"`
}

type IngredientResponse struct {
	Item   ItemResponse
	Amount int
}

type RecipeResponse struct {
	ID          uint
	Name        string
	Description string

	Ingredients []IngredientResponse

	Skill SkillResponse

	Duration Duration

	Amount int

	Item ItemResponse
}
