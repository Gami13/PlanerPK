import { createEffect } from 'solid-js';
import { useLocation, useNavigate } from '@solidjs/router';

export default function Authorize() {
  const location = useLocation();
  const navigate = useNavigate();

  createEffect(() => {
    const code = new URLSearchParams(location.search).get("code");

    if (code) {
      fetch(`http://localhost:8080/api/authorize?code=${code}`)
        .then(response => response.json())
        .then(data => {
          if (data.access_token) {
            document.cookie = `access_token=${data.access_token}; path=/; max-age=604800; samesite=strict`;
            navigate("/settings");
          }
        })
        .catch(error => {
          console.error("Authorization error:", error);
        });
    }
  });
  return (
    <main>
        <h1>Autoryzacja...</h1>
    </main>
  );
}
