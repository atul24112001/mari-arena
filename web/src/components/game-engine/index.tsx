import React, {
  useState,
  useEffect,
  useRef,
  ReactNode,
  CSSProperties,
  RefObject,
  forwardRef,
  useImperativeHandle,
} from "react";
import DefaultTimer from "./DefaultTimer";
import DefaultRenderer from "./DefaultRenderer";

type GameEngineProps = {
  initState?: Entities;
  initialState?: Entities;
  state?: Entities;
  initEntities?: Entities;
  initialEntities?: Entities;
  entities?: Entities | Promise<Entities>;
  systems: ((entities: Entities, args: any) => Entities)[];
  renderer?: (entities: Entities, window: Window) => ReactNode;
  running?: boolean;
  style?: CSSProperties;
  className?: string;
  onEvent?: (e: any) => void;
  children?: ReactNode;
  timer?: DefaultTimer;
};

const getEntitiesFromProps = (
  props: GameEngineProps
): Entities | Promise<Entities> =>
  props.initState ||
  props.initialState ||
  props.state ||
  props.initEntities ||
  props.initialEntities ||
  props.entities ||
  {};

const isPromise = (obj: any): obj is Promise<any> =>
  !!obj && typeof obj.then === "function";

const eventNames = `onClick onContextMenu onDoubleClick onDrag onDragEnd 
  onDragEnter onDragExit onDragLeave onDragOver 
  onDragStart onDrop onMouseDown onMouseEnter 
  onMouseLeave onMouseMove onMouseOut onMouseOver 
  onMouseUp onWheel onTouchCancel onTouchEnd 
  onTouchMove onTouchStart onKeyDown onKeyPress onKeyUp`;

const css = {
  container: {
    flex: 1,
    outline: "none",
  } as CSSProperties,
};

const GameEngine = forwardRef<GameEngineRef, GameEngineProps>((props, ref) => {
  const {
    systems = [],
    renderer = DefaultRenderer,
    running = true,
    timer = new DefaultTimer(),
    style,
    className,
    children,
    onEvent,
  } = props;
  //   const entities = props.entities
  const [entities, setEntities] = useState<Entities | null>(null);
  const input = useRef<any[]>([]);
  const speed = useRef<number>(2);
  const events = useRef<any[]>([]);
  const previousTime = useRef<number | null>(null);
  const previousDelta = useRef<number | null>(null);
  const container: RefObject<HTMLDivElement> = useRef(null);

  useEffect(() => {
    timer.subscribe(updateHandler);
    const loadEntities = async () => {
      let initialEntities = getEntitiesFromProps(props);
      if (isPromise(initialEntities)) initialEntities = await initialEntities;
      setEntities(initialEntities || {});
      if (running) start();
    };

    loadEntities();

    return () => {
      stop();
      timer.unsubscribe(updateHandler);
    };
  }, [running]);

  const clear = async () => {
    speed.current = 2;
    input.current.length = 0;
    events.current.length = 0;
    previousTime.current = null;
    previousDelta.current = null;
    if (props.initEntities) {
      setEntities(props.initEntities);
    }
  };

  const start = () => {
    clear();
    timer.start();
    dispatch({ type: "started" });
    container.current?.focus();
  };

  const stop = () => {
    timer.stop();
    dispatch({ type: "stopped" });
  };

  const swap = async (newEntities: Entities | Promise<Entities>) => {
    if (isPromise(newEntities)) newEntities = await newEntities;
    setEntities(newEntities || {});
    clear();
    dispatch({ type: "swapped" });
  };

  const dispatch = (e: any) => {
    setTimeout(() => {
      events.current.push(e);
      onEvent?.(e);
    }, 0);
  };

  const increaseSpeed = () => {
    speed.current += 0.1;
  };

  const updateHandler = (currentTime: number) => {
    const args = {
      input: input.current,
      window,
      events: events.current,
      dispatch,
      stop,
      speed: speed.current,
      increaseSpeed,
      time: {
        current: currentTime,
        previous: previousTime.current,
        delta: currentTime - (previousTime.current || currentTime),
        previousDelta: previousDelta.current,
      },
    };

    setEntities((prevEntities) => {
      const newEntities = systems.reduce(
        (state, system) => system(state, args),
        prevEntities || {}
      );
      input.current.length = 0;
      events.current.length = 0;
      previousTime.current = currentTime;
      previousDelta.current = args.time.delta;
      return newEntities;
    });
  };

  const inputHandlers = eventNames
    .split(/\s+/)
    .reduce((acc: { [key: string]: (e: any) => void }, eventName) => {
      acc[eventName] = (payload: any) => {
        payload.persist();
        input.current.push({ name: eventName, payload });
      };
      return acc;
    }, {});

  useImperativeHandle(ref, () => ({
    start,
    stop,
    swap,
    dispatch,
  }));
  return (
    <div
      ref={container}
      style={{ ...css.container, ...style }}
      className={className}
      tabIndex={0}
      {...inputHandlers}
    >
      {renderer(entities || {}, globalThis.window)}
      {children}
    </div>
  );
});

GameEngine.displayName = "GameEngine";
export default GameEngine;
