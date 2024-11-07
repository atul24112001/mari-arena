"use client";
import dynamic from "next/dynamic";
import Matter from "matter-js";

import { useEffect, useMemo, useRef, useState } from "react";
import Bird from "@/components/game/Bird";
import Physics from "@/components/game/Physics";
import Wall from "@/components/game/Wall";
import { Constants } from "@/lib/constants";
import Pipe, { generatePipe } from "@/components/game/Pipe";
import { useAuth } from "@/context/auth";
import { useRouter } from "next/navigation";
import { GameType } from "@prisma/client";
import { toast } from "sonner";
import Image from "next/image";

const GameEngine = dynamic(() => import("@/components/game-engine"), {
  ssr: false,
});

type Game = {
  isStarted: boolean;
  gameId: string;
  users: string[];
};

export default function GameClient({ gameType }: Props) {
  const gameEngine = useRef<GameEngineRef>(null);
  const { user, socket, setUser, sendMessage, underMaintenance } = useAuth();
  const router = useRouter();

  const [entities, setEntities] = useState<Entities>({});
  const [game, setGame] = useState<Game | null>(null);
  const [gameStartingIn, setGameStartingIn] = useState(0);
  const [engine] = useState<Matter.Engine>(
    Matter.Engine.create({ enableSleeping: false })
  );

  const [running, setRunning] = useState(false);
  const [gameOver, setGameOver] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  useEffect(() => {
    if (underMaintenance) {
      toast("Currently under maintenance, Please comeback after sometime");
      router.push("/");
    }
  }, [underMaintenance]);

  useEffect(() => {
    if (socket && confirmed) {
      sendMessage("join-random-game", {
        userId: user?.id,
        gameTypeId: gameType.id,
      });

      socket.onmessage = (e) => {
        const { type, data } = JSON.parse(e.data);

        if (type === "join-game") {
          setGame({
            gameId: data.gameId,
            isStarted: false,
            users: data.users,
          });
        } else if (type === "new-user") {
          setGame((prev) => ({
            gameId: data.gameId,
            isStarted: false,
            users: [...(prev?.users || []), data.userId],
          }));
        } else if (type === "start-game") {
          startGame();
        } else if (type === "error") {
          toast(data?.message || "Something went wrong", {
            duration: 2000,
          });
        } else if (type === "winner") {
          toast("You won!", {
            duration: 2000,
          });
          setUser((prev) => {
            if (!prev) {
              return prev;
            }
            return {
              ...prev,
              solanaBalance: prev.solanaBalance + data.amount,
            };
          });
          router.push("/");
        } else if (type === "loser") {
          toast("You lose!", {
            duration: 2000,
          });
          setUser((prev) => {
            if (!prev) {
              return prev;
            }
            return {
              ...prev,
              solanaBalance: prev.solanaBalance - data.amount,
            };
          });
          router.push("/");
        }
      };
    }
  }, [socket, confirmed]);

  const notEligibleToPlay = useMemo(() => {
    if (!user) {
      return "Please connect your wallet";
    }
    if (gameType.entry > user.solanaBalance) {
      return "Insufficient balance please add solana";
    }
    return null;
  }, [user, gameType]);

  useEffect(() => {
    if (notEligibleToPlay) {
      toast("Something went wrong", {
        description: notEligibleToPlay,
        className: "bg-primary text-red-500  border-0",
        descriptionClassName: "text-primary-foreground",
      });
      router.push("/");
    }
  }, [notEligibleToPlay]);

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
    if (!game || !user) {
      return;
    }
    if (e.type === "game-over") {
      setRunning(false);
      setGameOver(true);
      sendMessage("game-over", {
        gameId: game?.gameId,
        userId: user?.id,
      });
    } else if (e.type === "score") {
      sendMessage("update-board", {
        gameId: game?.gameId,
        userId: user?.id,
      });
    }
  };

  if (!user) {
    return null;
  }

  return (
    <div className="w-screen h-screen overflow-hidden relative z-50">
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
        <div className=" z-[1000] bg-[#00000024] fixed top-0 left-0 w-screen h-screen flex justify-center items-center">
          {gameOver ? (
            <div className="w-[90%] flex flex-col gap-1 items-center">
              <h2 className="font-bold text-primary-bg text-center">
                You can wait for results as other player(s) are/is still alive
                or you can go back home if you won your solana will be credited
                to your account
              </h2>
              <button
                onClick={() => router.push("/")}
                className="bg-yellow-100 text-black px-3 py-1"
              >
                Home
              </button>
            </div>
          ) : (
            <div className="w-[90%]">
              {gameStartingIn !== 0 ? (
                <p className="font-bold text-xl text-yellow-100">
                  Game starting in {gameStartingIn} sec...
                </p>
              ) : confirmed ? (
                <h2 className="mb-2 font-bold text-xl text-center text-yellow-100">
                  Waiting for opponents to join, pay attention game can start
                  anytime...
                </h2>
              ) : (
                <div className="flex flex-col items-center mb-2">
                  <button
                    onClick={() => setConfirmed(true)}
                    className="bg-yellow-100 text-black px-3 py-1"
                  >
                    Join game
                  </button>
                  <p className="text-bl">
                    One you join game you can&apos;t back-off, entry fees will
                    be deducted
                  </p>
                </div>
              )}
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
                    4. Whoever passes the most pipes will win.
                  </p>
                </article>
                <article>
                  <p className="font-semibold bg-yellow-100 px-2 py-1 rounded-sm mb-1 text-black">
                    5. Game is over when the bird touches anything.
                  </p>
                </article>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

type Props = {
  gameType: GameType;
};
