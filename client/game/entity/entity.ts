class Entity {
  private id: string;
  private components: Record<string, any>;

  constructor(id: string) {
    this.id = id;
    this.components = {};
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

  isEmpty() {
    return Object.keys(this.components).length === 0;
  }

  removeComponent(componentId: string) {
    delete this.components[componentId];
  }
}

export default Entity;
