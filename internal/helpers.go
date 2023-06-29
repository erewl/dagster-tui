package internal

import (
    "sort"
)

func Identity[T any](s T) T {
    return s
}

func Filter[T any](ss []T, cond func(T) bool) (ret []T) {
    for _, s := range ss {
        if cond(s) {
            ret = append(ret, s)
        }
    }
    return
}

func SortBy[T any](ss []T, c func(T) string) []T {
    sort.Slice(ss, func (i,j int) bool {
        return c(ss[i]) < c(ss[j])
    })
    return ss
}
