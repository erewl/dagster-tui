package app



func filter(ss []string, cond func(string) bool) (ret []string) {
    for _, s := range ss {
        if cond(s) {
            ret = append(ret, s)
        }
    }
    return
}