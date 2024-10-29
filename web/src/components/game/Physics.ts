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
    .filter(
      ({ name, payload }: any) =>
        name === "onClick" ||
        (name == "onKeyPress" && AllowedKeyCode[payload.code])
    )
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

          //   entities[`pipe${i}Top`].body.size = [
          //     Constants.PIPE_WIDTH,
          //     newPipeTopHeight,
          //   ];
          //   entities[`pipe${i}Bottom`].body.size = [
          //     Constants.PIPE_WIDTH,
          //     newPipeBottomHeight,
          //   ];

          //   Matter.World.remove(engine.world, topPipe.body);
          //   Matter.World.remove(engine.world, bottomPipe.body);
          //   const [newPipeTopHeight, newPipeBottomHeight] = generatePipe();

          //   const newPipeTop = Matter.Bodies.rectangle(
          //     lastBody.position.x + Constants.PIPE_GAP_SIZE,
          //     topPipe.body.position.y,
          //     Constants.PIPE_WIDTH,
          //     newPipeTopHeight
          //   );
          //   const newPipeBottom = Matter.Bodies.rectangle(
          //     lastBody.position.x + Constants.PIPE_GAP_SIZE,
          //     bottomPipe.body.position.y,
          //     Constants.PIPE_WIDTH,
          //     newPipeBottomHeight
          //   );

          //   entities[`pipe${i}Bottom`] = {
          //     body: newPipeBottom,
          //     color: "green",
          //     size: [Constants.PIPE_WIDTH, newPipeBottomHeight],
          //     renderer: Wall,
          //   };
          //   Matter.World.add(engine.world, [newPipeTop, newPipeBottom]);
        }
      }
    }
  }

  if (entities.bird.body.position.y > Constants.MAX_HEIGHT - 225) {
    entities.bird.body.position.y = Constants.MAX_HEIGHT - 225;
  } else if (entities.bird.body.position.y < 25) {
    dispatch({ type: "game-over" });
  }

  Matter.Engine.update(engine, time.delta);
  return { ...entities };
};

export default Physics;
