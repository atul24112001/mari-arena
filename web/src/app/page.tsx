import React from "react";
import ClientHomePage from "@/sections/home";
import { Card } from "@/components/helper";
import { GameType } from "@prisma/client";
import axios from "axios";
import AdminCreateGameType from "@/sections/home/admin/CreateGameType";

let cache = {
  gamesTypes: [] as GameType[],
  lastUpdated: new Date().getTime() - 1000,
};

export default async function Home({ searchParams }: ServerProps) {
  const { withoutCache } = await searchParams;
  let gamesTypes = cache.gamesTypes;
  if (cache.lastUpdated < new Date().getTime() || withoutCache) {
    console.log("Refetching");
    try {
      const { data } = await axios.get(
        `${process.env.NEXT_PUBLIC_API_URL}/api/game-types`
      );
      gamesTypes = data.data;
      cache.gamesTypes = data.data;
      cache.lastUpdated = new Date().getTime() + 60 * 60 * 1000;
    } catch (error) {
      console.log(error);
    }
  }

  return (
    <div className="h-screen bg-primary-bg overflow-hidden bg-cover md:bg-contain  ">
      <div className="px-4 md:px-10 py-4 mb-4 gap-4 flex justify-between items-center">
        <h1 className="text-primary text-sm md:text-lg   font-bold flex-1">
          Mari&nbsp;arena
        </h1>
        <ClientHomePage />
      </div>
      <div className="flex justify-center items-center">
        <div>
          <Card
            href="solo"
            title="Solo"
            description={`Entry Free | Winner 0 SOL`}
          />
          {gamesTypes.map((gamesType) => {
            let min = 1;
            if (gamesType.currency === "SOL") {
              min = 10 ** 9;
            }
            return (
              <Card
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
    </div>
  );
}

{
  /* bg-[url('/assets/background-day.png')] */
}
