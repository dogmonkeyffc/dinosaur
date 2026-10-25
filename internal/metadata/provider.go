package metadata

import "errors"

var ErrProviderAuth = errors.New("metadata provider rejected the credential")
var ErrProviderUnavailable = errors.New("metadata provider is unavailable")
var ErrProviderRateLimited = errors.New("metadata provider rate limit exceeded")
var ErrProviderNotFound = errors.New("movie does not exist on metadata provider")
