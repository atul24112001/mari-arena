import Matter from "matter-js";
import { Constants } from "@/lib/constants";
import Pipe, { generatePipe } from "./Pipe";
import Wall from "./Wall";

type Options = {
  touches: any;
  time: any;
  events: any[];
  input: any[];
  dispatch: (data: { type: string }) => void;
  speed: number;
  increaseSpeed: () => void;
};

const AllowedKeyCode: { [key: string]: boolean } = {
  KeyW: true,
};

const Physics = (
  entities: Entities,
  { time, input, dispatch, speed, increaseSpeed }: Options
) => {
  const engine: Matter.Engine = entities.physics.engine;
  const bird: Matter.Body = entities.bird.body;
  entities.bird.fly = false;
  input
    .filter(({ name, payload }: any) => {
      // console.log({ code: payload?.code });
      return (
        name === "onClick" ||
        (name == "onKeyPress" && AllowedKeyCode[payload.code])
      );
    })
    .forEach((e) => {
      entities.bird.fly = true;
      Matter.Body.applyForce(bird, bird.position, { x: 0.0, y: -0.08 });
    });

  for (let i = 1; i <= 10; i++) {
    const topPipe = entities[`pipe${i}Top`];
    const bottomPipe = entities[`pipe${i}Bottom`];
    if (topPipe && bottomPipe) {
      Matter.Body.translate(topPipe.body, {
        x: -speed,
        y: 0,
      });
      Matter.Body.translate(bottomPipe.body, {
        x: -speed,
        y: 0,
      });

      if (topPipe.body.position.x < -Constants.PIPE_WIDTH) {
        increaseSpeed();
        const lastBody = (
          i == 1 ? entities["pipe10Top"] : entities[`pipe${i - 1}Top`]
        )?.body;
        if (lastBody) {
          Matter.Body.translate(topPipe.body, {
            x:
              lastBody.position.x +
              Constants.PIPE_WIDTH / 2 +
              Constants.PIPE_GAP_SIZE,
            y: 0,
          });
          Matter.Body.translate(bottomPipe.body, {
            x:
              lastBody.position.x +
              Constants.PIPE_WIDTH / 2 +
              Constants.PIPE_GAP_SIZE,
            y: 0,
          });
        }
      }
    }
  }

  if (
    entities.bird.body.position.y < 25 ||
    entities.bird.body.position.x < 70 ||
    entities.bird.body.position.y >
      Constants.MAX_HEIGHT - Constants.FLOOR_HEIGHT ||
    entities.bird.body.position.x > Constants.MAX_WIDTH
  ) {
    dispatch({ type: "game-over" });
  }

  Matter.Engine.update(engine, time.delta);
  return { ...entities };
};

export default Physics;
