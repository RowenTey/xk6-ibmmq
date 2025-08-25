package ibmmq

import "github.com/grafana/sobek"

// freeze disallows resetting or changing the properties of the object.
func freeze(o *sobek.Object) {
	for _, key := range o.Keys() {
		if err := o.DefineDataProperty(
			key, o.Get(key), sobek.FLAG_FALSE, sobek.FLAG_FALSE, sobek.FLAG_TRUE); err != nil {
			panic(err)
		}
	}
}
