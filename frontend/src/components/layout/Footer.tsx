export function Footer() {
  return (
    <footer className="border-t bg-gray-50 py-4">
      <div className="container mx-auto px-4 text-center text-sm text-gray-600">
        © {new Date().getFullYear()} Sumo AI. All rights reserved.
      </div>
    </footer>
  );
}
