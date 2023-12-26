package models

type Food struct {
	Name string
}

func FoodsList() []Food {
	meals := []Food{
		{Name: "carrot salad"},
		{Name: "milk"},
		{Name: "energy drink"},
		{Name: "coffee"},
	}

	return meals
}
