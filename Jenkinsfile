
pipeline {
  agent any
  stages{
    stage('make clean')     { steps { sh 'make clean' } }
    stage('make all')       { steps { sh 'make all' } }
    stage('make test')      { steps { sh 'make test' } }
    stage('make benchmark') { steps { sh 'make benchmark' } }
  }
}

