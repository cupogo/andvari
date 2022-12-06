package comm

// DefaultModel struct contain model's default fields.
type DefaultModel struct {
	IDField      `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod    `bson:",inline"`
}

// Creating function call to it's inner fields defined hooks
func (model *DefaultModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *DefaultModel) Saving() error {
	return model.DateFields.Saving()
}

// DunceModel struct contain model's default fields with string pk.
type DunceModel struct {
	IDFieldStr   `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod    `bson:",inline"`
}

// SerialModel struct contain model's default fields.
type SerialModel struct {
	SerialField  `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod    `bson:",inline"`
}

// Creating function call to it's inner fields defined hooks
func (model *SerialModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *SerialModel) Saving() error {
	return model.DateFields.Saving()
}
