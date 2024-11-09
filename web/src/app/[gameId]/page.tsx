import React from "react";
import GameClient from "@/sections/game";
import { redirect } from "next/navigation";
import { GameType } from "@prisma/client";
import { getGameTypes } from "../page";

export default async function Game({ params }: ServerProps) {
  const { gameId } = await Promise.resolve(params);

  const gameTypes = await getGameTypes();
  const targetGameType = gameTypes.find(
    (gameType: GameType) => gameType.id === gameId
  );

  if (!targetGameType) {
    redirect("/");
  }

  return <GameClient gameType={targetGameType} />;
}
