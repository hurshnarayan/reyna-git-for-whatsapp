.PHONY: dev backend frontend bot build clean

# Start everything
dev:
	@echo "🤖 Starting Reyna..."
	@echo "   Backend  → http://localhost:8080"
	@echo "   Frontend → http://localhost:5173"
	@echo "   Bot      → scan QR in terminal"
	@make -j3 backend frontend bot

backend:
	cd backend && CGO_ENABLED=1 go run ./cmd/server/

frontend:
	cd frontend && npx vite --host 0.0.0.0

bot:
	cd whatsapp-bot && node bot.mjs

# Start without WhatsApp bot (dev mode)
dev-web:
	@make -j2 backend frontend

build:
	cd backend && CGO_ENABLED=1 go build -o bin/reyna-server ./cmd/server/
	cd frontend && npx vite build

clean:
	rm -rf backend/bin backend/reyna.db backend/drive_storage
	rm -rf frontend/dist frontend/node_modules/.vite
	rm -rf whatsapp-bot/auth_state whatsapp-bot/downloaded_files
