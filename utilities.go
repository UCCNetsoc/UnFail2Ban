package main

func filter(s [][]string, f func(string) bool) (ret [][]string) {
	for _, val := range s {
		if f(val[0]) {
			ret = append(ret, val)
		}
	}
	return
}
