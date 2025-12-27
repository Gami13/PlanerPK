import { PageTitle } from "../../Components/PageTitle";

export function NotFound() {
  return (
    <section class="text-gray-700 p-8">
      <PageTitle title="Nic tu nie ma :(" />
      <h1 >404: Not Found</h1>
      <p >It's gone 😞</p>
    </section>
  );
}
