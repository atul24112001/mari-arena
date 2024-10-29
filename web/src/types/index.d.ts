type Entity = { renderer: any; [key: string]: any };
type Entities = Record<string, Entity>;

type GameEngineRef = {
  start: () => void;
  stop: () => void;
  swap: (newEntities: Entities | Promise<Entities>) => Promise<void>;
  dispatch: (e: any) => void;
};
