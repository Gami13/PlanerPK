import { useLocation, useNavigate } from '@solidjs/router';
import { createEffect } from 'solid-js';
import { Endpoints } from '../../constants';
import { setAccessToken } from '../../utils';

export function Authorize() {
  const location = useLocation();
  const navigate = useNavigate();

  createEffect(() => {
    const code = new URLSearchParams(location.search).get("code");

    if (code) {
      fetch(`${Endpoints.Authorize}?code=${code}`)
        .then(response => response.json())
        .then(data => {
          if (data.access_token) {
            setAccessToken(data.access_token);
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
