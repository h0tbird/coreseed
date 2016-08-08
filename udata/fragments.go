package udata

//---------------------------------------------------------------------------
// func: loadFragments
//---------------------------------------------------------------------------

func (d *Data) loadFragments() {

	d.frags = append(d.frags, frag{
		filter: filter{
			anyOf:  []string{"a", "b"},
			noneOf: []string{"c", "d"},
			allOf:  []string{"e", "f"},
		},
		data: "Hello",
	})

	d.frags = append(d.frags, frag{
		filter: filter{
			anyOf:  []string{"a", "b"},
			noneOf: []string{"c", "d"},
			allOf:  []string{"e", "f"},
		},
		data: " world",
	})
}
