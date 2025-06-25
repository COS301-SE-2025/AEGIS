//verifies whether a user is authenticated (e.g., by checking the presence of a valid JWT in localStorage or Zustand).
export function isAuthenticated(): boolean {
  const token = localStorage.getItem('token');
  return !!token;
}
