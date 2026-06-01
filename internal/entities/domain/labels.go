package domain

import "unicode/utf8"

const maxLabelLength = 10

type Label struct {
	id     int64
	scanID int64
	name   string
}

func (l Label) Name() string {
	return l.name
}

func (l Label) ID() int64 {
	return l.id
}

func (l Label) ScanID() int64 {
	return l.scanID
}

func NewLabel(name string) (Label, error) {
	if utf8.RuneCountInString(name) == 0 {
		return Label{}, NewError("Label", "label cannot be empty")
	}

	if utf8.RuneCountInString(name) > maxLabelLength {
		return Label{}, NewError("Label", "label cannot be longer than 10 characters")
	}
	return Label{
		name: name,
	}, nil
}

func ReconstituteLabel(id int64, scanID int64, name string) Label {
	return Label{
		id:     id,
		scanID: scanID,
		name:   name,
	}
}
