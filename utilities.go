package main

func filter(s []string, f func(string) bool) (ret []string) {
	for _, val := range s {
		if f(val) {
			ret = append(ret, val)
		}
	}
	return
}

func mapf(s []string, f func(string) string) []string {
	fs := make([]string, len(s))
	for i, val := range s {
		fs[i] = f(val)
	}
	return fs
}

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
