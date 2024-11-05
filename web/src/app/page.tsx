import React from "react";
import ClientHomePage from "@/sections/home";
import { Card } from "@/components/helper";
import { GameType } from "@prisma/client";
import axios from "axios";
import AdminCreateGameType from "@/sections/home/admin/CreateGameType";
import { StarsBackground } from "@/components/ui/star-background";
import { ShootingStars } from "@/components/ui/shooting-stars";

// export let cache = {
//   gamesTypes: [] as GameType[],
//   lastUpdated: new Date().getTime() - 1000,
// };

export default async function Home({ searchParams }: ServerProps) {
  const { withoutCache } = await Promise.resolve(searchParams);

  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/game-types`,
    { cache: withoutCache ? "no-cache" : "force-cache" }
  );
  const { data } = await response.json();

  return (
    <>
      <div className="relative z-10 px-4 md:px-10 py-4 mb-4 gap-4 flex justify-between items-center">
        <h1 className="text-primary text-sm md:text-lg   font-bold flex-1">
          Mari&nbsp;arena
        </h1>
        <ClientHomePage />
      </div>
      <div className="relative z-10 flex justify-center items-center">
        <div>
          <Card
            href="solo"
            title="Solo"
            description={`Entry Free | Winner 0 SOL`}
          />
          {data.map((gamesType: GameType) => {
            let min = 1;
            if (gamesType.currency === "SOL") {
              min = 10 ** 9;
            }
            return (
              <Card
                key={gamesType.id}
                href={gamesType.id}
                title={gamesType.title}
                description={`Entry ${gamesType.entry / min} ${
                  gamesType.currency
                } | Winner ${gamesType.winner / min} ${gamesType.currency}`}
              />
            );
          })}
          <AdminCreateGameType />
        </div>
      </div>
    </>
  );
}
