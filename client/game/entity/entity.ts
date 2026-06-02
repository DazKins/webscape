class Entity {
  private id: string;
  private components: Record<string, any>;
  private availableInteractions: string[];

  constructor(id: string) {
    this.id = id;
    this.components = {};
    this.availableInteractions = [];
  }

  getId() {
    return this.id;
  }

  updateComponent(componentId: string, data: any) {
    this.components[componentId] = data;
  }

  getComponent(componentId: string) {
    return this.components[componentId];
  }

  setAvailableInteractions(availableInteractions: string[]) {
    this.availableInteractions = availableInteractions;
  }

  getAvailableInteractions() {
    return this.availableInteractions;
  }

  isEmpty() {
    return Object.keys(this.components).length === 0;
  }

  removeComponent(componentId: string) {
    delete this.components[componentId];
  }
}

export default Entity;
