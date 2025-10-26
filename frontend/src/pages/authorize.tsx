import { createEffect, Suspense } from 'solid-js';
import AboutData from './about.data';

export default function About() {
  const name = AboutData();

  createEffect(() => {
    console.log(name());
  });

  return (
    <section class="bg-pink-100 text-gray-700 p-8">
        <h1>Authorization</h1>
    </section>
  );
}
