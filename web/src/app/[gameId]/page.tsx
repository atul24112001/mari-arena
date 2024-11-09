import React from "react";
import GameClient from "@/sections/game";
import { redirect } from "next/navigation";
import { GameType } from "@prisma/client";
import { customCache } from "../page";

export default async function Game({ params }: ServerProps) {
  const { gameId } = await Promise.resolve(params);

  let gameTypes: GameType[] = customCache.gameTypes;
  if (customCache.lastUpdated < new Date().getTime()) {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/api/game-types`
    );
    const data = await response.json();
    customCache.lastUpdated = new Date().getTime() + 3600000;
    customCache.gameTypes = data.data;
    gameTypes = data.data;
  }

  const targetGameType = gameTypes.find(
    (gameType: GameType) => gameType.id === gameId
  );

  if (!targetGameType) {
    redirect("/");
  }

  return <GameClient gameType={targetGameType} />;
}
