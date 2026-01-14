# Sumo AI Frontend

RAG-based AI Assistant built with Next.js 14, TypeScript, and Tailwind CSS.

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
```bash
# Update .env.local with your API Gateway URL
NEXT_PUBLIC_API_URL=https://your-api-gateway-url.amazonaws.com
```

3. Run development server:
```bash
npm run dev
```

4. Open [http://localhost:3000](http://localhost:3000)

## Features

- **Authentication**: Login with Bolt Database
- **Chat Interface**: Real-time chat with AI assistant
- **Session Management**: Create and manage multiple chat sessions
- **Document Ingestion**: Upload documents to knowledge base
- **RAG Support**: Context-aware responses using uploaded documents

## Tech Stack

- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS
- shadcn/ui components
- Axios for API calls
- React Context for state management

## Project Structure

```
src/
├── app/              # Next.js pages
├── components/       # React components
├── context/          # React context providers
├── hooks/            # Custom React hooks
├── lib/              # Utilities and API client
└── styles/           # Global styles
```

## Build

```bash
npm run build
npm start
```
