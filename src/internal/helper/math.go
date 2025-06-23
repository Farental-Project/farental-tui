package helper

func calc10Inc(n int) int {
	r := n % 10

	if r == 0 {
		return 10
	}

	return r
}

func Next10Inc(n int) int {
	r := calc10Inc(n)

	if r < 10 {
		return 10 - r
	}

	return r
}

func Prev10Inc(n int) int {
	r := calc10Inc(n)

	return r
}
