package cache

// Global cache that can be shared across the whole program.
var Global = Scoped(DefaultExpiration)
