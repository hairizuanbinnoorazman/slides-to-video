format:
	elm-format --yes src/

gen:
	elm make --optimize --output=main.js ./src/App.elm
	@if command -v uglifyjs >/dev/null 2>&1; then \
		uglifyjs main.js --compress "pure_funcs=[F2,F3,F4,F5,F6,F7,F8,F9,A2,A3,A4,A5,A6,A7,A8,A9],pure_getters,keep_fargs=false,unsafe_comps,unsafe" | uglifyjs --mangle --output main.min.js && \
		rm main.js; \
	else \
		echo "uglifyjs not found, skipping minification"; \
		mv main.js main.min.js; \
	fi