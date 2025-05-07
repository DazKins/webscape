export default function addReferenceGeometry(scene) {
  const axisLength = 1;
  const axisThickness = 0.05;

  const xAxisGeometry = new THREE.BoxGeometry(
    axisLength,
    axisThickness,
    axisThickness
  );
  const xAxisMaterial = new THREE.MeshBasicMaterial({ color: 0xff0000 });
  const xAxis = new THREE.Mesh(xAxisGeometry, xAxisMaterial);
  xAxis.position.set(axisLength / 2, 0, 0);
  scene.add(xAxis);

  const zAxisGeometry = new THREE.BoxGeometry(
    axisThickness,
    axisThickness,
    axisLength
  );
  const zAxisMaterial = new THREE.MeshBasicMaterial({ color: 0x0000ff });
  const zAxis = new THREE.Mesh(zAxisGeometry, zAxisMaterial);
  zAxis.position.set(0, 0, axisLength / 2);
  scene.add(zAxis);

  const referenceGeometry = new THREE.SphereGeometry(0.2, 16, 16);
  const referenceMaterial = new THREE.MeshBasicMaterial({ color: 0xff0000 });
  const referencePoint = new THREE.Mesh(referenceGeometry, referenceMaterial);
  referencePoint.position.set(0, 0, 0);
  scene.add(referencePoint);
}
