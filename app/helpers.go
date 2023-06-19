package app


func filter[T any](ss []T, cond func(T) bool) (ret []T) {
    for _, s := range ss {
        if cond(s) {
            ret = append(ret, s)
        }
    }
    return
}
