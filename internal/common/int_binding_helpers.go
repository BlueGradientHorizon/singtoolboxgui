package common

import "fyne.io/fyne/v2/data/binding"

// Increments value of [binding.Int]
func BindingInc(b binding.Int) {
	v, err := b.Get()
	if err == nil {
		b.Set(v + 1)
	}
}

// Decrements value of [binding.Int]
func BindingDec(b binding.Int) {
	v, err := b.Get()
	if err == nil {
		b.Set(v - 1)
	}
}

// Adds v to value of [binding.Int]
func BindingAdd(b binding.Int, v int) {
	vOld, err := b.Get()
	if err == nil {
		b.Set(vOld + v)
	}
}
