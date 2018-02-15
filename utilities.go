package main

func filter(s []string, f func(string) bool) (ret []string) {
	for _, val := range s {
		if f(val) {
			ret = append(ret, val)
		}
	}
	return
}

func mapf(s []string, f func(string) string) {
	for i, val := range s {
		if len(val) < 1 {
			continue
		}
		s[i] = f(val)
	}
}

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
