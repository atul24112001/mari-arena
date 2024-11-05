type Entity = { [key: string]: any };
type Entities = Record<string, Entity>;

type GameEngineRef = {
  start: () => void;
  stop: () => void;
  swap: (newEntities: Entities | Promise<Entities>) => Promise<void>;
  dispatch: (e: { type: string }) => void;
};

type Params = { [key: string]: string };
type SearchParams = { [key: string]: string | string[] | undefined };

type ServerProps = {
  params: Promise<Params>;
  searchParams: Promise<SearchParams>;
};
