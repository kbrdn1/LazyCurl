# GitHub Labels Documentation

This document presents a proposed organization of labels for our GitHub project, as well as priorities and issue types for our organization. This structure aims to improve the classification and tracking of our tickets.

## Table of Contents
- [GitHub Labels](#github-labels)
  - [Type Labels](#type-labels)
  - [Domain Labels](#domain-labels)
  - [Management Labels](#management-labels)
  - [Status Labels](#status-labels)
- [Priorities](#priorities)
- [Issue Types](#issue-types)

## GitHub Labels

### Type Labels

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **feature** | `#0E8A16` (Green) | Implementation of a new feature | Label pour identifier l'ajout d'une nouvelle fonctionnalité au projet |
| **fix** | `#D73A4A` (Light red) | Bug fix | Correction d'un problème existant dans le code |
| **hotfix** | `#FF3333` (Bright red) | Critical bug fix in production | Correction urgente d'un bug critique en environnement de production |
| **docs** | `#1D76DB` (Blue) | Improvements or additions to documentation | Modifications ou ajouts concernant la documentation technique ou utilisateur |
| **test** | `#87CEEB` (Sky blue) | Unit, integration, or functional tests | Ajout ou modification de tests automatisés |
| **refactor** | `#FBCA04` (Yellow) | Code restructuring without changing external behavior | Amélioration de la structure du code sans modifier son comportement externe |
| **chore** | `#808080` (Gray) | Routine maintenance tasks not modifying the main codebase (configuration, scripts, etc.) | Tâches de maintenance qui ne modifient pas directement le code principal (configuration, scripts, etc.) |
| **optimisation** | `#FFA500` (Orange) | Performance or logic improvements | Améliorations des performances ou de la logique du code |

### Domain Labels

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **authentication** | `#0075CA` (Dark blue) | Related to authentication | Concerne les fonctionnalités d'authentification et de gestion des sessions |
| **errors** | `#D73A4A` (Red) | Related to error handling | Concerne la gestion des erreurs et exceptions |
| **ci/cd** | `#26A69A` (Teal) | Continuous integration / Continuous deployment | Modifications liées à l'intégration continue ou au déploiement continu |
| **security** | `#B60205` (Dark red) | Security issues to fix | Problèmes de sécurité à corriger |
| **UI/UX** | `#FF69B4` (Pink) | User interface or user experience improvements | Améliorations de l'interface utilisateur ou de l'expérience utilisateur |

### Management Labels

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **dependencies** | `#8B008B` (Purple) | Updates or issues with dependencies | Mises à jour ou problèmes avec les dépendances du projet |
| **breaking** | `#FF0000` (Red) | Non-backward compatible changes | Changements qui ne sont pas rétrocompatibles |
| **versioning** | `#000080` (Navy blue) | Version management | Gestion des versions du projet |
| **urgent** | `#FF1493` (Deep pink) | Requires immediate attention | Nécessite une attention immédiate |

### Status Labels

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **duplicate** | `#CCCCCC` (Gray) | Duplicated issue or PR | Issue ou Pull Request qui duplique une autre déjà existante |
| **invalid** | `#444444` (Dark gray) | Invalid issue that cannot be resolved | Issue non valide ou qui ne peut pas être résolue |
| **wontfix** | `#FFFFFF` (White) | Valid issue that will not be fixed | Issue valide mais qui ne sera pas corrigée pour une raison spécifique |

## Priorities (for Github Project)

Priorities are managed via a dedicated selector and not via labels to avoid redundancy.

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **Critical** | Red | Blocking issue that must be resolved immediately. Severely affects core functionality. | Problème bloquant qui doit être résolu immédiatement. Affecte gravement les fonctionnalités principales. |
| **High** | Orange | Important issue that is not blocking but needs to be resolved quickly. Significant impact on user experience. | Problème important qui n'est pas bloquant mais doit être résolu rapidement. Impact significatif sur l'expérience utilisateur. |
| **Medium** | Yellow | Standard priority issue that should be addressed in the current development cycle. | Problème de priorité standard qui devrait être traité dans le cycle de développement actuel. |
| **Low** | Green | Minor issue that can be deferred to future development cycles. | Problème mineur qui peut être reporté aux prochains cycles de développement. |
| **Trivial** | Blue | Cosmetic or non-essential improvements that can be handled when time permits. | Améliorations cosmétiques ou non essentielles qui peuvent être traitées lorsque le temps le permet. |

## Issue Types (for Organisation)

| Name | Color | Description (EN) | Note (FR) |
|-----|---------|-----------------|-----------| 
| **Task** | Blue | Specific task to accomplish, usually small to medium-sized. | Tâche spécifique à accomplir, généralement de petite à moyenne taille. |
| **Story** | Green | Feature described from the end user's perspective. | Fonctionnalité décrite du point de vue de l'utilisateur final. |
| **Bug** | Red | Problem or incorrect behavior to fix. | Problème ou comportement incorrect à corriger. |
| **Epic** | Purple | Set of stories or tasks grouped for a major feature. | Ensemble de stories ou de tâches regroupées pour une fonctionnalité majeure. |
| **Question** | Blue | Request for clarification or information. | Demande de clarification ou d'information. |
| **Improvement** | Blue | Enhancement of an existing feature. | Amélioration d'une fonctionnalité existante. |
| **Research** | Yellow | Preliminary investigation to explore solutions. | Investigation préliminaire pour explorer des solutions. |
| **Infrastructure** | Gray | Changes concerning infrastructure or development environment. | Modifications concernant l'infrastructure ou l'environnement de développement. |
| **Support Request** | Orange | Request for assistance requiring intervention. | Demande d'assistance nécessitant une intervention. |
| **Technical Documentation** | Blue | Internal technical documentation for the development team. | Documentation technique interne pour l'équipe de développement. |
| **Spike** | Pink | Time-limited technical investigation to assess feasibility. | Investigation technique limitée dans le temps pour évaluer la faisabilité. |
| **Incident** | Red | Problem that occurred in production requiring immediate intervention. | Problème survenu en production nécessitant une intervention immédiate. |
| **Feature** | Green | Request for a new feature to be implemented | Demande d'implémentation d'une nouvelle fonctionnalité |

---

## Implementation Instructions

To implement these labels in GitHub, you can:

1. Access `https://github.com/[organisation]/[repository]/labels`
2. Delete existing unwanted labels
3. Create new labels with the names, descriptions, and colors indicated above
4. For more efficient implementation, you can use tools like [GitHub Label Sync](https://github.com/Financial-Times/github-label-sync)

To configure issue types and priorities, you will need to create appropriate issue templates in the `.github/ISSUE_TEMPLATE/` folder of your repository.
