# MusicStream Frontend

A modern, Spotify-inspired music streaming frontend built with Next.js 14, React, TypeScript, and Tailwind CSS.

## Features

- 🎵 Browse artists, albums, and songs
- 🔍 Search for music, artists, and albums
- 📚 Manage your library and playlists
- ▶️ Audio player with queue management
- 🎨 Dark theme with smooth animations
- 🔐 User authentication (login/register)

## Tech Stack

- **Framework:** Next.js 14 (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **State Management:** Zustand
- **HTTP Client:** Axios
- **Icons:** Lucide React
- **Notifications:** React Hot Toast

## Project Structure

```
frontend/
├── app/                    # Next.js App Router pages
│   ├── page.tsx           # Home page
│   ├── login/             # Login/register page
│   ├── search/            # Search page
│   ├── library/           # User library
│   ├── artists/[id]/      # Artist detail page
│   └── albums/[id]/       # Album detail page
├── components/            # React components
│   ├── Sidebar.tsx        # Navigation sidebar
│   └── Player.tsx         # Audio player
├── lib/                   # Utilities and helpers
│   ├── api.ts            # API client
│   ├── types.ts          # TypeScript types
│   └── store.ts          # Zustand stores
└── public/               # Static assets
```

## Development

### Prerequisites

- Node.js 20+
- npm or yarn

### Setup

1. Install dependencies:
```bash
npm install
```

2. Set environment variables:
```bash
# Create .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_URL_BROWSER=http://localhost:8080
```

3. Run development server:
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

### Build for Production

```bash
npm run build
npm start
```

## Docker

Build and run with Docker:

```bash
docker build -t musicstream-frontend .
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://gateway-api:8080 \
  -e NEXT_PUBLIC_API_URL_BROWSER=http://localhost:8080 \
  musicstream-frontend
```

## API Integration

The frontend connects to the gateway API at port 8080. The API client handles:

- Authentication (JWT tokens stored in cookies)
- Automatic token injection in requests
- User, artist, album, and music endpoints
- Search functionality
- Playlist management

## Environment Variables

- `NEXT_PUBLIC_API_URL`: API URL for server-side requests (e.g., `http://gateway-api:8080`)
- `NEXT_PUBLIC_API_URL_BROWSER`: API URL for client-side requests (e.g., `http://localhost:8080`)

## License

MIT
