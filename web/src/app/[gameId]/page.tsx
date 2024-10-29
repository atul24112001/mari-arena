"use client";
import dynamic from "next/dynamic";
import Matter from "matter-js";

import { useContext, useEffect, useRef, useState } from "react";
import Bird from "@/components/game/Bird";
import Physics from "@/components/game/Physics";
import Wall from "@/components/game/Wall";
import { Constants } from "@/lib/constants";
import Pipe, { generatePipe } from "@/components/game/Pipe";
import { useAuth } from "@/context/auth";
import { useRouter } from "next/navigation";
import { useWalletModal } from "@solana/wallet-adapter-react-ui";
const GameEngine = dynamic(() => import("@/components/game-engine"), {
  ssr: false,
});

export default function Home() {
  const gameEngine = useRef<GameEngineRef>(null);
  const { user } = useAuth();
  const router = useRouter();
  const walletModal = useWalletModal();

  const [gameStartingIn, setGameStartingIn] = useState(0);
  const [entities, setEntities] = useState<Entities>({});
  const [running, setRunning] = useState(false);
  const [engine] = useState<Matter.Engine>(
    Matter.Engine.create({ enableSleeping: false })
  );

  useEffect(() => {
    if (!user) {
      walletModal.setVisible(true);
      router.push("/");
    }
  }, [user]);

  const getPipes = () => {
    const pipes: {
      pipeTop: Matter.Body;
      pipeBottom: Matter.Body;
      entities: Entities;
    }[] = [];
    for (let i = 1; i <= 10; i++) {
      const [pipeTopHeight, pipeBottomHeight] = generatePipe();
      const pipeTop = Matter.Bodies.rectangle(
        Constants.MAX_WIDTH -
          Constants.PIPE_WIDTH / 2 +
          (Constants.PIPE_GAP_SIZE * i - 1),
        pipeTopHeight / 2,
        Constants.PIPE_WIDTH,
        pipeTopHeight,
        { isStatic: true, density: 1000, friction: 1, id: i }
      );
      const pipeBottom = Matter.Bodies.rectangle(
        Constants.MAX_WIDTH -
          Constants.PIPE_WIDTH / 2 +
          (Constants.PIPE_GAP_SIZE * i - 1),
        Constants.MAX_HEIGHT - pipeBottomHeight / 2,
        Constants.PIPE_WIDTH,
        pipeBottomHeight,
        { isStatic: true, density: 1000, friction: 1, id: i + 1 }
      );
      pipes.push({
        pipeTop,
        pipeBottom,
        entities: {
          [`pipe${i}Top`]: {
            body: pipeTop,
            color: "green",
            size: [Constants.PIPE_WIDTH, pipeTopHeight],
            renderer: Pipe,
            top: true,
          },
          [`pipe${i}Bottom`]: {
            body: pipeBottom,
            color: "green",
            size: [Constants.PIPE_WIDTH, pipeBottomHeight],
            renderer: Pipe,
          },
        },
      });
    }

    return pipes;
  };

  const startGame = () => {
    setGameStartingIn(10);
    const interval = setInterval(() => {
      setGameStartingIn((prev) => prev - 1);
    }, 1000);
    setTimeout(() => {
      clearInterval(interval);
      gameEngine.current?.swap(entities);
      setRunning(true);
    }, 10000);
  };

  useEffect(() => {
    if (user) {
      const world = engine.world;
      const bird = Matter.Bodies.rectangle(
        Constants.MAX_WIDTH / 4,
        Constants.MAX_HEIGHT / 3,
        70,
        50
      );
      const floor = Matter.Bodies.rectangle(
        Constants.MAX_WIDTH / 2,
        Constants.MAX_HEIGHT - Constants.FLOOR_HEIGHT / 2,
        Constants.MAX_WIDTH,
        Constants.FLOOR_HEIGHT,
        { isStatic: true, density: 100 }
      );

      const pipes = getPipes();
      Matter.World.add(world, [
        bird,
        // ceiling,

        ...pipes.reduce((pipes, current) => {
          pipes.push(current.pipeTop);
          pipes.push(current.pipeBottom);
          return pipes;
        }, [] as Matter.Body[]),
        floor,
      ]);

      const pipeEntities = pipes.reduce((prev, curr) => {
        return { ...prev, ...curr.entities };
      }, {} as Entities);

      Matter.Events.on(engine, "collisionStart", (a) => {
        gameEngine.current?.dispatch({ type: "game-over" });
      });

      setEntities({
        physics: {
          engine,
          world,
          renderer: undefined,
        },
        bird: { body: bird, color: "green", size: [70, 50], renderer: Bird },
        ...pipeEntities,
        floor: {
          body: floor,
          color: "#ff5252",
          size: [Constants.MAX_WIDTH, Constants.FLOOR_HEIGHT],
          renderer: Wall,
        },
      });
    }
  }, [user]);

  const onEvent = (e: any) => {
    if (e.type === "game-over") {
      setRunning(false);
    }
  };

  if (!user) {
    return null;
  }

  return (
    <div className="w-screen h-screen overflow-hidden">
      <GameEngine
        ref={gameEngine}
        running={running}
        systems={[Physics]}
        initEntities={entities}
        onEvent={onEvent}
        className=""
      >
        <div className="h-screen bg-cover md:bg-contain overflow-hidden bg-[url('/assets/background-day.png')]"></div>
      </GameEngine>
      {!running && (
        <div className="bg-[#00000024] fixed z-50 top-0 left-0 w-screen h-screen flex justify-center items-center">
          {gameStartingIn !== 0 ? (
            <p className="font-bold text-xl">
              Game starting in {gameStartingIn} sec...
            </p>
          ) : (
            <button
              onClick={startGame}
              className="bg-yellow-300 text-black px-3 py-1"
            >
              Start game
            </button>
          )}
        </div>
      )}
    </div>
  );
}
