"use client";
import dynamic from "next/dynamic";
import Matter from "matter-js";

import { useEffect, useRef, useState } from "react";
import Bird from "@/components/game/Bird";
import Physics from "@/components/game/Physics";
import Wall from "@/components/game/Wall";
import { Constants } from "@/lib/constants";
import Pipe, { generatePipe } from "@/components/game/Pipe";
import Image from "next/image";
import { useRouter } from "next/navigation";
const GameEngine = dynamic(() => import("@/components/game-engine"), {
  ssr: false,
});

export default function Home() {
  const gameEngine = useRef<GameEngineRef>(null);

  const [gameStartingIn, setGameStartingIn] = useState(0);
  const [entities, setEntities] = useState<Entities>({});
  const [running, setRunning] = useState(false);
  const router = useRouter();
  const [engine] = useState<Matter.Engine>(
    Matter.Engine.create({ enableSleeping: false })
  );

  const [gameOver, setGameOver] = useState(false);

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
    const world = engine.world;
    const bird = Matter.Bodies.rectangle(
      Constants.MAX_WIDTH / 4,
      Constants.MAX_HEIGHT / 3,
      70,
      50
      // { density: 0.1 }
    );
    // const ceiling = Matter.Bodies.rectangle(
    //   Constants.MAX_WIDTH / 2,
    //   Constants.CEILING_HEIGHT / 2,
    //   Constants.MAX_WIDTH,
    //   Constants.CEILING_HEIGHT,
    //   { isStatic: true, density: 100 }
    // );
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

    Matter.Events.on(engine, "collisionStart", () => {
      gameEngine.current?.dispatch({ type: "game-over" });
    });

    setEntities({
      physics: {
        engine,
        world,
        renderer: undefined,
      },
      bird: { body: bird, color: "green", size: [70, 50], renderer: Bird },
      // ceiling: {
      //   body: ceiling,
      //   color: "#ff5252",
      //   size: [Constants.MAX_WIDTH, 30],
      //   renderer: Wall,
      // },
      ...pipeEntities,
      floor: {
        body: floor,
        color: "#ff5252",
        size: [Constants.MAX_WIDTH, Constants.FLOOR_HEIGHT],
        renderer: Wall,
      },
    });
  }, []);

  const onEvent = (e: any) => {
    if (e.type === "game-over") {
      setRunning(false);
      setGameOver(true);
    }
  };

  return (
    <div className="h-screen w-screen overflow-clip relative z-50">
      <GameEngine
        ref={gameEngine}
        running={running}
        systems={[Physics]}
        initEntities={entities}
        onEvent={onEvent}
      >
        <div className="h-screen  bg-cover md:bg-contain overflow-hidden md bg-[url('/assets/background-day.png')]"></div>
      </GameEngine>

      {!running && (
        <div className=" z-[1000] bg-[#00000024] fixed top-0 left-0 w-screen h-screen flex justify-center items-center">
          {gameOver ? (
            <div className="flex flex-col items-center font-bold">
              <p className="text-yellow-100  bg-red-500  px-3 py-1 mb-2">
                Game Over
              </p>
              <button
                onClick={() => router.push("/")}
                className="bg-yellow-100 text-black px-3 py-1"
              >
                Home
              </button>
            </div>
          ) : (
            <div className="w-[90%] max-w-96">
              <div>
                <h3 className="font-bold mb-1 text-lg text-yellow-100">
                  Rules
                </h3>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    1. Once the game is started you can&apos;t opt out if you
                    close the tab you will be considered dead with 0 points and
                    sol won&apos;t be refunded
                  </p>
                </article>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    2. This is a dangerous area(image) for birds if they come
                    into contact with the pipe. Although the bird appears to be
                    far away, it is actually not.
                  </p>
                  <Image
                    width={300}
                    height={100}
                    alt="rule-2"
                    className="mx-auto w-full h-auto mb-2"
                    src="/assets/rules/rule-2.png"
                  />
                </article>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    3. The speed of the bird will increase gradually.
                  </p>
                </article>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    4. Whoever passes the most pipes will win
                  </p>
                </article>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    4. Game is over when the bird touches anything.
                  </p>
                </article>
              </div>
              {gameStartingIn !== 0 ? (
                <p className="font-bold text-xl text-center text-yellow-100 mt-3">
                  Game starting in {gameStartingIn} sec...
                </p>
              ) : (
                <div className="flex justify-center items-center mt-3">
                  <button
                    onClick={startGame}
                    className="bg-yellow-100  text-black px-3 py-1"
                  >
                    Start game
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
