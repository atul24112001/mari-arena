import React from "react";
import GameClient from "@/sections/game";
import { redirect } from "next/navigation";
import { GameType } from "@prisma/client";

export default async function Game({ params }: ServerProps) {
  const { gameId } = await Promise.resolve(params);

  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/game-type`
  );
  const { data } = await response.json();

  const targetGameType = data.find(
    (gameType: GameType) => gameType.id === gameId
  );

  if (!targetGameType) {
    redirect("/");
  }

  return <GameClient gameType={targetGameType} />;
}
