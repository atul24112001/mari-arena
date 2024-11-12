import React from "react";
import ClientHomePage from "@/sections/home";
import { Card } from "@/components/helper";
import { GameType } from "@prisma/client";
import AdminCreateGameType from "@/sections/home/admin/CreateGameType";
import Error500 from "@/sections/error/Error500";

export const customCache = {
  lastUpdated: new Date().getTime() - 1000,
  gameTypes: [] as GameType[],
};

export const getGameTypes = async () => {
  let gameTypes: GameType[] = customCache.gameTypes;
  if (customCache.lastUpdated < new Date().getTime()) {
    console.log("Updating game types");
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/api/game-types`,
      { cache: "no-cache" }
    );
    if (response.status) {
      const data = await response.json();
      gameTypes = data.data;
      customCache.lastUpdated = new Date().getTime() + 3600000;
      customCache.gameTypes = data.data;
    } else {
      throw Error("Something went wrong");
    }
  }
  return gameTypes;
};

export default async function Home() {
  try {
    const gameTypes = await getGameTypes();

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
            {gameTypes.map((gamesType: GameType) => {
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
  } catch (error) {
    console.log(error);
  }

  return <Error500 />;
}
