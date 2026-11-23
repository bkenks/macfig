package main

import (
	"fmt"
	"strconv"
	"strings"
)

// maxPerKind guards against typos like `--small 500`.
const maxPerKind = 100

// countValue is a pflag.Value holding a non-negative count, or "all" (-1)
// when allowAll is set.
type countValue struct {
	n        int
	allowAll bool
}

const countAll = -1

func (c *countValue) String() string {
	if c.n == countAll {
		return "all"
	}
	return strconv.Itoa(c.n)
}

func (c *countValue) Set(s string) error {
	if c.allowAll && strings.EqualFold(s, "all") {
		c.n = countAll
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		if c.allowAll {
			return fmt.Errorf("want a number or \"all\", got %q", s)
		}
		return fmt.Errorf("want a number, got %q", s)
	}
	if n > maxPerKind {
		return fmt.Errorf("%d is more than the limit of %d", n, maxPerKind)
	}
	c.n = n
	return nil
}

func (c *countValue) Type() string {
	if c.allowAll {
		return "N|all"
	}
	return "N"
}
