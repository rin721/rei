package service

// RefreshTokenCacheKey 返回用户刷新令牌在缓存中的键名。
func RefreshTokenCacheKey(userID string) string {
	return refreshTokenCachePrefix + userID
}
