package config

func AllowedOrigins() []string {
	if IsDevelopment() {
		return []string{"*"}
	}

	return []string{
		"https://vatusa.net",
	}
}
