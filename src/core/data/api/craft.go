package api

type CraftStartBody struct {
	RecipeID uint `validate:"required"`
	Amount   int  `validate:"required,number"`
}

type RecipeResponse struct {
	ID          uint
	Name        string
	Description string

	Ingredients []IngredientResponse

	Skill SkillResponse

	Duration DurationResponse

	Amount uint

	Item ItemResponse
}

type IngredientResponse struct {
	Item   ItemResponse
	Amount uint
}
