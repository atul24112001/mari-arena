export default class DefaultTimer {
  private subscribers: ((time: number) => void)[] = [];
  private loopId: number | null = null;
  //   private intervalId: NodeJS.Timeout | null = null;

  private loop = (time?: number) => {
    if (this.loopId) {
      if (time) {
        this.subscribers.forEach((callback) => callback(time));
      }
    }
    this.loopId = requestAnimationFrame(this.loop);
  };

  start() {
    if (this.loopId === null) {
      this.loop();
    }
  }

  stop() {
    if (this.loopId !== null) {
      cancelAnimationFrame(this.loopId);
      this.loopId = null;
    }
  }

  subscribe(callback: (time: number) => void) {
    if (!this.subscribers.includes(callback)) {
      this.subscribers.push(callback);
    }
  }

  unsubscribe(callback: (time: number) => void) {
    this.subscribers = this.subscribers.filter((sub) => sub !== callback);
  }
}
